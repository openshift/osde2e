package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	junit "github.com/joshdk/go-junit"
	"golang.org/x/net/html"
	"k8s.io/apimachinery/pkg/util/wait"
	restclient "k8s.io/client-go/rest"
)

var (
	resultsPortStr = strconv.Itoa(resultsPort)

	errNotRun = errors.New("suite has not run yet")
)

func ensurePassingXML(results map[string][]byte) (hadXML bool, err error) {
	// ensure the junit xml indicates a passing job
	var match bool
	for filename, data := range results {
		log.Println("checking", filename)
		match, err = filepath.Match("junit*.xml", filename)
		if err != nil {
			err = fmt.Errorf("Failed matching filename %s: %w", filename, err)
			return
		}
		if match {
			hadXML = true
			// Use Ginkgo's JUnitTestSuite to unmarshal the JUnit XML file
			suites, e := junit.Ingest(data)
			if e != nil {
				err = fmt.Errorf("Failed parsing junit xml in %s: %w", filename, e)
				return
			}
			for _, suite := range suites {
				for _, testcase := range suite.Tests {
					if (testcase.Error) != nil {
						err = fmt.Errorf("at least one test failed (see junit xml for more): %s", testcase.Error)
						return
					}
				}
			}
		}
	}
	return
}

// RetrieveResults gathers the results from the test Pod. Should only be called after tests are finished.
func (r *Runner) RetrieveResults() (map[string][]byte, error) {
	results, err := r.retrieveResultsForDirectory("")
	if err != nil {
		return nil, fmt.Errorf("failed retrieving results: %w", err)
	}
	return results, err
}

// RetrieveTestResults gathers and validates the results from the test Pod. Should only be called after tests are finished. This method both fetches the results and ensures that they contain valid JUnit XML indicating that all tests passed.
func (r *Runner) RetrieveTestResults() (map[string][]byte, error) {
	results, err := r.RetrieveResults()
	if err != nil {
		return nil, fmt.Errorf("failed retrieving results: %w", err)
	}
	hadXML, err := ensurePassingXML(results)
	if err != nil {
		return results, fmt.Errorf("failed checking results for Junit XML report: %w", err)
	}
	if !hadXML {
		return results, fmt.Errorf("results did not contain Junit XML report")
	}
	return results, err
}

func (r *Runner) retrieveResultsForDirectory(directory string) (map[string][]byte, error) {
	var rdr io.ReadCloser
	var resp restclient.ResponseWrapper
	var err error
	if r.svc == nil {
		return nil, errNotRun
	}

	// request result list
	// sometimes it is possible for the service/endpoint to not be ready before the results are finished.
	// we loop through here five times with a sleep statement to check.
	wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		resp = r.Kube.CoreV1().Services(r.Namespace).ProxyGet("http", r.svc.Name, resultsPortStr, directory, nil)
		rdr, err = resp.Stream(context.TODO())
		if err != nil {
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not retrieve result file listing: %v", err)
	}

	// parse list
	n, err := html.Parse(rdr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse result file listing: %v", err)
	} else if err = rdr.Close(); err != nil {
		return nil, err
	}

	// download each file
	results := map[string][]byte{}
	if err = r.downloadLinks(n, results, directory); err != nil {
		return results, fmt.Errorf("encountered error downloading results: %v", err)
	}
	return results, nil
}

// downloadLinks, given an html page, will download all present links.
// This is useful when a pod is publishing an html list of artifacts.
func (r *Runner) downloadLinks(n *html.Node, results map[string][]byte, directory string) error {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" {
				if strings.HasSuffix(a.Val, "/") {
					var newDirectory string
					if directory != "" {
						newDirectory = a.Val
					} else {
						newDirectory = path.Join(directory, a.Val)
					}

					log.Println("Downloading directory " + newDirectory)
					directoryResults, err := r.retrieveResultsForDirectory(newDirectory)
					if err != nil {
						log.Printf("error while getting results for directory %s: %v", newDirectory, err)
						continue
					}

					for k, v := range directoryResults {
						results[k] = v
					}
				} else {
					resp := r.Kube.CoreV1().Services(r.Namespace).ProxyGet("http", r.svc.Name, resultsPortStr, path.Join(directory, a.Val), nil)
					data, err := resp.DoRaw(context.TODO())
					if err != nil {
						return err
					}

					filename := a.Val
					log.Println("Downloading " + filename)
					results[path.Join(directory, filename)] = data
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := r.downloadLinks(c, results, directory); err != nil {
			log.Printf("error while getting results for %s: %v", directory, err)
		}
	}
	return nil
}
