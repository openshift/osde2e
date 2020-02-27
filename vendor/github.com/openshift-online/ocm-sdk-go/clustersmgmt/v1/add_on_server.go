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

// AddOnServer represents the interface the manages the 'add_on' resource.
type AddOnServer interface {

	// Delete handles a request for the 'delete' method.
	//
	// Deletes the add-on.
	Delete(ctx context.Context, request *AddOnDeleteServerRequest, response *AddOnDeleteServerResponse) error

	// Get handles a request for the 'get' method.
	//
	// Retrieves the details of the add-on.
	Get(ctx context.Context, request *AddOnGetServerRequest, response *AddOnGetServerResponse) error

	// Update handles a request for the 'update' method.
	//
	// Updates the add-on.
	Update(ctx context.Context, request *AddOnUpdateServerRequest, response *AddOnUpdateServerResponse) error
}

// AddOnDeleteServerRequest is the request for the 'delete' method.
type AddOnDeleteServerRequest struct {
}

// AddOnDeleteServerResponse is the response for the 'delete' method.
type AddOnDeleteServerResponse struct {
	status int
	err    *errors.Error
}

// Status sets the status code.
func (r *AddOnDeleteServerResponse) Status(value int) *AddOnDeleteServerResponse {
	r.status = value
	return r
}

// AddOnGetServerRequest is the request for the 'get' method.
type AddOnGetServerRequest struct {
}

// AddOnGetServerResponse is the response for the 'get' method.
type AddOnGetServerResponse struct {
	status int
	err    *errors.Error
	body   *AddOn
}

// Body sets the value of the 'body' parameter.
//
//
func (r *AddOnGetServerResponse) Body(value *AddOn) *AddOnGetServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *AddOnGetServerResponse) Status(value int) *AddOnGetServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'get' method.
func (r *AddOnGetServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// AddOnUpdateServerRequest is the request for the 'update' method.
type AddOnUpdateServerRequest struct {
	body *AddOn
}

// Body returns the value of the 'body' parameter.
//
//
func (r *AddOnUpdateServerRequest) Body() *AddOn {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
//
func (r *AddOnUpdateServerRequest) GetBody() (value *AddOn, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal request to the
// 'update' method.
func (r *AddOnUpdateServerRequest) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(addOnData)
	err = decoder.Decode(data)
	if err != nil {
		return err
	}
	r.body, err = data.unwrap()
	if err != nil {
		return err
	}
	return err
}

// AddOnUpdateServerResponse is the response for the 'update' method.
type AddOnUpdateServerResponse struct {
	status int
	err    *errors.Error
	body   *AddOn
}

// Body sets the value of the 'body' parameter.
//
//
func (r *AddOnUpdateServerResponse) Body(value *AddOn) *AddOnUpdateServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *AddOnUpdateServerResponse) Status(value int) *AddOnUpdateServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'update' method.
func (r *AddOnUpdateServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// dispatchAddOn navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchAddOn(w http.ResponseWriter, r *http.Request, server AddOnServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "DELETE":
			adaptAddOnDeleteRequest(w, r, server)
		case "GET":
			adaptAddOnGetRequest(w, r, server)
		case "PATCH":
			adaptAddOnUpdateRequest(w, r, server)
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

// readAddOnDeleteRequest reads the given HTTP requests and translates it
// into an object of type AddOnDeleteServerRequest.
func readAddOnDeleteRequest(r *http.Request) (*AddOnDeleteServerRequest, error) {
	var err error
	result := new(AddOnDeleteServerRequest)
	return result, err
}

// writeAddOnDeleteResponse translates the given request object into an
// HTTP response.
func writeAddOnDeleteResponse(w http.ResponseWriter, r *AddOnDeleteServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	return nil
}

// adaptAddOnDeleteRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptAddOnDeleteRequest(w http.ResponseWriter, r *http.Request, server AddOnServer) {
	request, err := readAddOnDeleteRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(AddOnDeleteServerResponse)
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
	err = writeAddOnDeleteResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// readAddOnGetRequest reads the given HTTP requests and translates it
// into an object of type AddOnGetServerRequest.
func readAddOnGetRequest(r *http.Request) (*AddOnGetServerRequest, error) {
	var err error
	result := new(AddOnGetServerRequest)
	return result, err
}

// writeAddOnGetResponse translates the given request object into an
// HTTP response.
func writeAddOnGetResponse(w http.ResponseWriter, r *AddOnGetServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptAddOnGetRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptAddOnGetRequest(w http.ResponseWriter, r *http.Request, server AddOnServer) {
	request, err := readAddOnGetRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(AddOnGetServerResponse)
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
	err = writeAddOnGetResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// readAddOnUpdateRequest reads the given HTTP requests and translates it
// into an object of type AddOnUpdateServerRequest.
func readAddOnUpdateRequest(r *http.Request) (*AddOnUpdateServerRequest, error) {
	var err error
	result := new(AddOnUpdateServerRequest)
	err = result.unmarshal(r.Body)
	if err != nil {
		return nil, err
	}
	return result, err
}

// writeAddOnUpdateResponse translates the given request object into an
// HTTP response.
func writeAddOnUpdateResponse(w http.ResponseWriter, r *AddOnUpdateServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptAddOnUpdateRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptAddOnUpdateRequest(w http.ResponseWriter, r *http.Request, server AddOnServer) {
	request, err := readAddOnUpdateRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(AddOnUpdateServerResponse)
	response.status = 204
	err = server.Update(r.Context(), request, response)
	if err != nil {
		glog.Errorf(
			"Can't process request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	err = writeAddOnUpdateResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
