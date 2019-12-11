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

package v1 // github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/golang/glog"
	"github.com/openshift-online/ocm-sdk-go/errors"
)

// LogServer represents the interface the manages the 'log' resource.
type LogServer interface {

	// Get handles a request for the 'get' method.
	//
	// Retrieves the details of the log.
	Get(ctx context.Context, request *LogGetServerRequest, response *LogGetServerResponse) error
}

// LogGetServerRequest is the request for the 'get' method.
type LogGetServerRequest struct {
}

// LogGetServerResponse is the response for the 'get' method.
type LogGetServerResponse struct {
	status int
	err    *errors.Error
	body   *Log
}

// Body sets the value of the 'body' parameter.
//
//
func (r *LogGetServerResponse) Body(value *Log) *LogGetServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *LogGetServerResponse) Status(value int) *LogGetServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'get' method.
func (r *LogGetServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// dispatchLog navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchLog(w http.ResponseWriter, r *http.Request, server LogServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "GET":
			adaptLogGetRequest(w, r, server)
		default:
			errors.SendMethodNotAllowed(w, r)
			return
		}
	} else {
		switch segments[0] {
		default:
			errors.SendNotFound(w, r)
			return
		}
	}
}

// readLogGetRequest reads the given HTTP requests and translates it
// into an object of type LogGetServerRequest.
func readLogGetRequest(r *http.Request) (*LogGetServerRequest, error) {
	var err error
	result := new(LogGetServerRequest)
	return result, err
}

// writeLogGetResponse translates the given request object into an
// HTTP response.
func writeLogGetResponse(w http.ResponseWriter, r *LogGetServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptLogGetRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptLogGetRequest(w http.ResponseWriter, r *http.Request, server LogServer) {
	request, err := readLogGetRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(LogGetServerResponse)
	response.status = 200
	err = server.Get(r.Context(), request, response)
	if err != nil {
		glog.Errorf(
			"Can't process request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	err = writeLogGetResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
