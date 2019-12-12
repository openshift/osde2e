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

package v1 // github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/golang/glog"
	"github.com/openshift-online/ocm-sdk-go/errors"
)

// RegistryServer represents the interface the manages the 'registry' resource.
type RegistryServer interface {

	// Get handles a request for the 'get' method.
	//
	// Retrieves the details of the registry.
	Get(ctx context.Context, request *RegistryGetServerRequest, response *RegistryGetServerResponse) error
}

// RegistryGetServerRequest is the request for the 'get' method.
type RegistryGetServerRequest struct {
}

// RegistryGetServerResponse is the response for the 'get' method.
type RegistryGetServerResponse struct {
	status int
	err    *errors.Error
	body   *Registry
}

// Body sets the value of the 'body' parameter.
//
//
func (r *RegistryGetServerResponse) Body(value *Registry) *RegistryGetServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *RegistryGetServerResponse) Status(value int) *RegistryGetServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'get' method.
func (r *RegistryGetServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// dispatchRegistry navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchRegistry(w http.ResponseWriter, r *http.Request, server RegistryServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "GET":
			adaptRegistryGetRequest(w, r, server)
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

// readRegistryGetRequest reads the given HTTP requests and translates it
// into an object of type RegistryGetServerRequest.
func readRegistryGetRequest(r *http.Request) (*RegistryGetServerRequest, error) {
	var err error
	result := new(RegistryGetServerRequest)
	return result, err
}

// writeRegistryGetResponse translates the given request object into an
// HTTP response.
func writeRegistryGetResponse(w http.ResponseWriter, r *RegistryGetServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptRegistryGetRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptRegistryGetRequest(w http.ResponseWriter, r *http.Request, server RegistryServer) {
	request, err := readRegistryGetRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(RegistryGetServerResponse)
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
	err = writeRegistryGetResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
