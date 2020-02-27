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

// AddOnInstallationServer represents the interface the manages the 'add_on_installation' resource.
type AddOnInstallationServer interface {

	// Delete handles a request for the 'delete' method.
	//
	// Deletes the add-on installation.
	Delete(ctx context.Context, request *AddOnInstallationDeleteServerRequest, response *AddOnInstallationDeleteServerResponse) error

	// Get handles a request for the 'get' method.
	//
	// Retrieves the details of the add-on installation.
	Get(ctx context.Context, request *AddOnInstallationGetServerRequest, response *AddOnInstallationGetServerResponse) error
}

// AddOnInstallationDeleteServerRequest is the request for the 'delete' method.
type AddOnInstallationDeleteServerRequest struct {
}

// AddOnInstallationDeleteServerResponse is the response for the 'delete' method.
type AddOnInstallationDeleteServerResponse struct {
	status int
	err    *errors.Error
}

// Status sets the status code.
func (r *AddOnInstallationDeleteServerResponse) Status(value int) *AddOnInstallationDeleteServerResponse {
	r.status = value
	return r
}

// AddOnInstallationGetServerRequest is the request for the 'get' method.
type AddOnInstallationGetServerRequest struct {
}

// AddOnInstallationGetServerResponse is the response for the 'get' method.
type AddOnInstallationGetServerResponse struct {
	status int
	err    *errors.Error
	body   *AddOnInstallation
}

// Body sets the value of the 'body' parameter.
//
//
func (r *AddOnInstallationGetServerResponse) Body(value *AddOnInstallation) *AddOnInstallationGetServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *AddOnInstallationGetServerResponse) Status(value int) *AddOnInstallationGetServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'get' method.
func (r *AddOnInstallationGetServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// dispatchAddOnInstallation navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchAddOnInstallation(w http.ResponseWriter, r *http.Request, server AddOnInstallationServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "DELETE":
			adaptAddOnInstallationDeleteRequest(w, r, server)
		case "GET":
			adaptAddOnInstallationGetRequest(w, r, server)
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

// readAddOnInstallationDeleteRequest reads the given HTTP requests and translates it
// into an object of type AddOnInstallationDeleteServerRequest.
func readAddOnInstallationDeleteRequest(r *http.Request) (*AddOnInstallationDeleteServerRequest, error) {
	var err error
	result := new(AddOnInstallationDeleteServerRequest)
	return result, err
}

// writeAddOnInstallationDeleteResponse translates the given request object into an
// HTTP response.
func writeAddOnInstallationDeleteResponse(w http.ResponseWriter, r *AddOnInstallationDeleteServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	return nil
}

// adaptAddOnInstallationDeleteRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptAddOnInstallationDeleteRequest(w http.ResponseWriter, r *http.Request, server AddOnInstallationServer) {
	request, err := readAddOnInstallationDeleteRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(AddOnInstallationDeleteServerResponse)
	response.status = 204
	err = server.Delete(r.Context(), request, response)
	if err != nil {
		glog.Errorf(
			"Can't process request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	err = writeAddOnInstallationDeleteResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// readAddOnInstallationGetRequest reads the given HTTP requests and translates it
// into an object of type AddOnInstallationGetServerRequest.
func readAddOnInstallationGetRequest(r *http.Request) (*AddOnInstallationGetServerRequest, error) {
	var err error
	result := new(AddOnInstallationGetServerRequest)
	return result, err
}

// writeAddOnInstallationGetResponse translates the given request object into an
// HTTP response.
func writeAddOnInstallationGetResponse(w http.ResponseWriter, r *AddOnInstallationGetServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptAddOnInstallationGetRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptAddOnInstallationGetRequest(w http.ResponseWriter, r *http.Request, server AddOnInstallationServer) {
	request, err := readAddOnInstallationGetRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(AddOnInstallationGetServerResponse)
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
	err = writeAddOnInstallationGetResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
