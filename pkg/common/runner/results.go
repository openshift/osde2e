package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"path"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"k8s.io/apimachinery/pkg/util/wait"
	restclient "k8s.io/client-go/rest"
)

var (
	resultsPortStr = strconv.Itoa(resultsPort)

	errNotRun = errors.New("suite has not run yet")
)

// RetrieveResults gathers the results from the test Pod. Should only be called after tests are finished.
func (r *Runner) RetrieveResults() (map[string][]byte, error) {
	return r.retrieveResultsForDirectory("")
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
						return fmt.Errorf("error while getting results for directory %s: %v", newDirectory, err)
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
			return err
		}
	}
	return nil
}
