/*
Copyright (c) 2018 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This file contains the implementation of the methods of the connection that are used to send HTTP
// requests and receive HTTP responses.

package sdk

import (
	"context"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	strip "github.com/grokify/html-strip-tags-go"
)

var wsRegex = regexp.MustCompile(`\s+`)

func (c *Connection) RoundTrip(request *http.Request) (response *http.Response, err error) {
	// Check if the connection is closed:
	err = c.checkClosed()
	if err != nil {
		return
	}

	// Get the context from the request:
	ctx := request.Context()

	// Get and delete the header that contains the anonymized path that should be used to
	// report metrics:
	var metric string
	header := request.Header
	if header != nil {
		metric = header.Get(metricHeader)
		header.Del(metricHeader)
	}
	if metric == "" {
		metric = "/-"
	}

	// Measure the time that it takes to send the request and receive the response:
	// it in the log:
	before := time.Now()
	response, err = c.send(ctx, request)
	after := time.Now()
	elapsed := after.Sub(before)

	// Update the metrics:
	if c.callCountMetric != nil || c.callDurationMetric != nil {
		code := 0
		if response != nil {
			code = response.StatusCode
		}
		labels := map[string]string{
			metricsAPIServiceLabel: c.GetAPIServiceLabelFromPath(request.URL.Path),
			metricsMethodLabel:     request.Method,
			metricsPathLabel:       metric,
			metricsCodeLabel:       strconv.Itoa(code),
		}
		if c.callCountMetric != nil {
			c.callCountMetric.With(labels).Inc()
		}
		if c.callDurationMetric != nil {
			c.callDurationMetric.With(labels).Observe(elapsed.Seconds())
		}
	}

	return
}

func (c *Connection) send(ctx context.Context, request *http.Request) (response *http.Response,
	err error) {
	// Check that the request URL:
	if request.URL.Path == "" {
		err = fmt.Errorf("request path is mandatory")
		return
	}
	if request.URL.Scheme != "" || request.URL.Host != "" || !path.IsAbs(request.URL.Path) {
		err = fmt.Errorf("request URL '%s' isn't absolute", request.URL)
		return
	}

	// Add the API URL to the request URL:
	request.URL = c.apiURL.ResolveReference(request.URL)

	// Check the request method and body:
	switch request.Method {
	case http.MethodGet, http.MethodDelete:
		if request.Body != nil {
			err = fmt.Errorf(
				"request body is not allowed for the '%s' method",
				request.Method,
			)
			return
		}
	case http.MethodPost, http.MethodPatch, http.MethodPut:
		// POST and PATCH and PUT don't need to have a body. It is up to the server to decide if
		// this is acceptable.
	default:
		err = fmt.Errorf("method '%s' is not allowed", request.Method)
		return
	}

	// Get the access token:
	token, _, err := c.TokensContext(ctx)
	if err != nil {
		err = fmt.Errorf("can't get access token: %v", err)
		return
	}

	// Add the default headers:
	if request.Header == nil {
		request.Header = make(http.Header)
	}
	if c.agent != "" {
		request.Header.Set("User-Agent", c.agent)
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	switch request.Method {
	case http.MethodPost, http.MethodPatch, http.MethodPut:
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("Accept", "application/json")

	// Send the request and get the response:
	response, err = c.client.Do(request)
	if err != nil {
		err = fmt.Errorf("can't send request: %v", err)
		return
	}

	// Check that the response content type is JSON:
	err = c.checkContentType(response)
	if err != nil {
		return
	}

	return
}

// checkContentType checks that the content type of the given response is JSON. Note that if the
// content type isn't JSON this method will consume the complete body in order to generate an error
// message containing a summary of the content.
func (c *Connection) checkContentType(response *http.Response) error {
	var err error
	var mediaType string
	contentType := response.Header.Get("Content-Type")
	if contentType != "" {
		mediaType, _, err = mime.ParseMediaType(contentType)
		if err != nil {
			return err
		}
	} else {
		mediaType = contentType
	}
	if !strings.EqualFold(mediaType, "application/json") {
		var summary string
		summary, err = c.contentSummary(mediaType, response)
		if err != nil {
			return fmt.Errorf(
				"expected response content type 'application/json' but received "+
					"'%s' and couldn't obtain content summary: %v",
				mediaType, err,
			)
		}
		return fmt.Errorf(
			"expected response content type 'application/json' but received '%s' and "+
				"content '%s'",
			mediaType, summary,
		)
	}
	return nil
}

// contentSummary reads the body of the given response and returns a summary it. The summary will
// be the complete body if it isn't too log. If it is too long then the summary will be the
// beginning of the content followed by ellipsis.
func (c *Connection) contentSummary(mediaType string, response *http.Response) (summary string, err error) {
	var body []byte
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	limit := 200
	runes := []rune(string(body))
	if strings.EqualFold(mediaType, "text/html") && len(runes) > limit {
		content := strip.StripTags(string(body))
		content = wsRegex.ReplaceAllString(strings.TrimSpace(content), " ")
		runes = []rune(content)
	}
	if len(runes) > limit {
		summary = fmt.Sprintf("%s...", string(runes[:200]))
	} else {
		summary = string(runes)
	}
	return
}
