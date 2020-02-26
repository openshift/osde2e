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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

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
			metricsMethodLabel: request.Method,
			metricsPathLabel:   metric,
			metricsCodeLabel:   strconv.Itoa(code),
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
	case http.MethodPost, http.MethodPatch:
		// POST and PATCH don't need to have a body. It is up to the server to decide if
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
	case http.MethodPost, http.MethodPatch:
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("Accept", "application/json")

	// If debug is enabled then we need to read the complete body in memory, in order to send it
	// to the log, and we need to replace the original with a reader that reads it from memory:
	if c.logger.DebugEnabled() {
		if request.Body != nil {
			var body []byte
			body, err = ioutil.ReadAll(request.Body)
			if err != nil {
				err = fmt.Errorf("can't read request body: %v", err)
				return
			}
			err = request.Body.Close()
			if err != nil {
				err = fmt.Errorf("can't close request body: %v", err)
				return
			}
			c.dumpRequest(ctx, request, body)
			request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		} else {
			c.dumpRequest(ctx, request, nil)
		}
	}

	// Send the request and get the response:
	response, err = c.client.Do(request)
	if err != nil {
		err = fmt.Errorf("can't send request: %v", err)
		return
	}

	// If debug is enabled then we need to read the complete response body in memory, in order
	// to send it the log, and we need to replace the original with a reader that reads it from
	// memory:
	if c.logger.DebugEnabled() {
		if response.Body != nil {
			var body []byte
			body, err = ioutil.ReadAll(response.Body)
			if err != nil {
				err = fmt.Errorf("can't read response body: %v", err)
				return
			}
			err = response.Body.Close()
			if err != nil {
				err = fmt.Errorf("can't close response body: %v", err)
				return
			}
			c.dumpResponse(ctx, response, body)
			response.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		} else {
			c.dumpResponse(ctx, response, nil)
		}
	}

	// check if json
	if !strings.EqualFold(response.Header.Get("Content-Type"), "application/json") {
		err = fmt.Errorf("expected JSON content type but received '%s'; %s",
			response.Header.Get("Content-Type"),
			getResponseInfo(response))
		return
	}

	return
}

func getResponseInfo(response *http.Response) string {
	info := fmt.Sprintf("request: %s %s; response status: %d %s",
		response.Request.Method,
		response.Request.URL,
		response.StatusCode,
		response.Status,
	)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		info = fmt.Sprintf("%s; can't read response body: %v", info, err)
		return info
	}
	err = response.Body.Close()
	if err != nil {
		info = fmt.Sprintf("%s; can't close response body: %v", info, err)
		return info
	}
	response.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	bodyStr := []rune(string(body))
	if len(bodyStr) > 200 {
		bodyStr = bodyStr[:200]
	}
	info = fmt.Sprintf("%s; response body: %s", info, string(bodyStr))
	return info
}
