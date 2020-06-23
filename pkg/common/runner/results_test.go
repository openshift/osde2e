package runner

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	kubev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	kubetest "k8s.io/client-go/testing"
)

var (
	expectedResults = map[string][]byte{
		"a": []byte("testdata"),
		"b": []byte("moretestdata"),
		"c": []byte("evenmoretestdata"),
	}

	resultPage = response(`
<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 3.2 Final//EN"><html>
<title>Directory listing for /</title>
<body>
<h2>Directory listing for /</h2>
<hr>
<ul>
<li><a href="a">a</a>
<li><a href="b">b</a>
<li><a href="c">c</a>
</ul>
<hr>
</body>
</html>
`)
)

func TestRetrieveResults(t *testing.T) {
	// setup mock client
	client := fake.NewSimpleClientset()
	client.AddProxyReactor("services", resultsServerReactor)

	// setup runner
	def := *DefaultRunner
	r := &def
	r.Kube = client

	// create results service
	svc, err := r.createService(new(kubev1.Pod))
	if err != nil {
		t.Fatalf("Failed to create example service: %v", err)
	}
	r.svc = svc

	// get results
	results, err := r.RetrieveResults()
	if err != nil {
		t.Fatalf("Failed to get results: %v", err)
	}

	// compare to expected
	for k, v := range expectedResults {
		if actualV, ok := results[k]; !ok {
			t.Fatalf("missing file '%s' in results", k)
		} else if !bytes.Equal(actualV, v) {
			t.Fatalf("file '%s' has been corrupted: want '%s', got '%s'", k, v, actualV)
		}
	}
}

func resultsServerReactor(action kubetest.Action) (handled bool, ret rest.ResponseWrapper, err error) {
	proxyAction := action.(kubetest.ProxyGetActionImpl)

	// only respond to service proxy requests
	if handled = proxyAction.Matches(http.MethodGet, "services"); !handled {
		return
	}

	path := strings.TrimPrefix(proxyAction.Path, "/")
	if path == "" {
		ret = resultPage
	} else if data, ok := expectedResults[path]; ok {
		ret = response(data)
	} else {
		ret = response{}
	}

	return
}

type response []byte

func (r response) DoRaw(context.Context) ([]byte, error) {
	if len(r) == 0 {
		return nil, errors.New("file does not exist")
	}
	return r, nil
}
func (r response) Stream(context.Context) (io.ReadCloser, error) {
	if len(r) == 0 {
		return nil, errors.New("file does not exist")
	}
	return ioutil.NopCloser(bytes.NewReader(r)), nil
}
