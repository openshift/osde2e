/*
Copyright (c) 2019 Red Hat, Inc.

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

// IMPORTANT: This file has been generated automatically, refrain from modifying it manually as all
// your changes will be lost when the file is generated again.

package helpers // github.com/openshift-online/ocm-sdk-go/helpers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// AddValue creates the given set of query parameters if needed, an then adds
// the given parameter.
func AddValue(query *url.Values, name string, value interface{}) {
	if *query == nil {
		*query = make(url.Values)
	}
	query.Add(name, fmt.Sprintf("%v", value))
}

// CopyQuery creates a copy of the given set of query parameters.
func CopyQuery(query url.Values) url.Values {
	if query == nil {
		return nil
	}
	result := make(url.Values)
	for name, values := range query {
		result[name] = CopyValues(values)
	}
	return result
}

// AddHeader creates the given set of headers if needed, and then adds the given
// header:
func AddHeader(header *http.Header, name string, value interface{}) {
	if *header == nil {
		*header = make(http.Header)
	}
	header.Add(name, fmt.Sprintf("%v", value))
}

// SetHeader creates a copy of the given set of headers, and adds the header
// containing the given metrics path.
func SetHeader(header http.Header, metric string) http.Header {
	result := make(http.Header)
	for name, values := range header {
		result[name] = CopyValues(values)
	}
	result.Set(metricHeader, metric)
	return result
}

// CopyValues copies a slice of strings.
func CopyValues(values []string) []string {
	if values == nil {
		return nil
	}
	result := make([]string, len(values))
	copy(result, values)
	return result
}

// Segments calculates the path segments for the given path.
func Segments(path string) []string {
	for strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	for strings.HasSuffix(path, "/") {
		path = path[0 : len(path)-1]
	}
	return strings.Split(path, "/")
}

// Name of the header used to contain the metrics path:
const metricHeader = "X-Metric"
