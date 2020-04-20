package runner

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"golang.org/x/net/html"
	"k8s.io/apimachinery/pkg/util/wait"
	restclient "k8s.io/client-go/rest"
)

var (
	resultsPortStr = strconv.Itoa(resultsPort)

	ErrNotRun = errors.New("suite has not run yet")
)

// RetrieveResults gathers the results from the test Pod. Should only be called after tests are finished.
func (r *Runner) RetrieveResults() (map[string][]byte, error) {
	var rdr io.ReadCloser
	var resp restclient.ResponseWrapper
	var err error
	if r.svc == nil {
		return nil, ErrNotRun
	}

	// request result list
	// sometimes it is possible for the service/endpoint to not be ready before the results are finished.
	// we loop through here five times with a sleep statement to check.
	wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		resp = r.Kube.CoreV1().Services(r.Namespace).ProxyGet("http", r.svc.Name, resultsPortStr, "/", nil)
		rdr, err = resp.Stream()
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
	if err = r.downloadLinks(n, results); err != nil {
		return results, fmt.Errorf("encountered error downloading results: %v", err)
	}
	return results, nil
}

// downloadLinks, given an html page, will download all present links.
// This is useful when a pod is publishing an html list of artifacts.
func (r *Runner) downloadLinks(n *html.Node, results map[string][]byte) error {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" {
				resp := r.Kube.CoreV1().Services(r.Namespace).ProxyGet("http", r.svc.Name, resultsPortStr, "/"+a.Val, nil)
				data, err := resp.DoRaw()
				if err != nil {
					return err
				}

				filename := a.Val
				log.Println("Downloading " + filename)
				results[filename] = data
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := r.downloadLinks(c, results); err != nil {
			return err
		}
	}
	return nil
}
