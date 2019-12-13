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

// RoleServer represents the interface the manages the 'role' resource.
type RoleServer interface {

	// Delete handles a request for the 'delete' method.
	//
	// Deletes the role.
	Delete(ctx context.Context, request *RoleDeleteServerRequest, response *RoleDeleteServerResponse) error

	// Get handles a request for the 'get' method.
	//
	// Retrieves the details of the role.
	Get(ctx context.Context, request *RoleGetServerRequest, response *RoleGetServerResponse) error

	// Update handles a request for the 'update' method.
	//
	// Updates the role.
	Update(ctx context.Context, request *RoleUpdateServerRequest, response *RoleUpdateServerResponse) error
}

// RoleDeleteServerRequest is the request for the 'delete' method.
type RoleDeleteServerRequest struct {
}

// RoleDeleteServerResponse is the response for the 'delete' method.
type RoleDeleteServerResponse struct {
	status int
	err    *errors.Error
}

// Status sets the status code.
func (r *RoleDeleteServerResponse) Status(value int) *RoleDeleteServerResponse {
	r.status = value
	return r
}

// RoleGetServerRequest is the request for the 'get' method.
type RoleGetServerRequest struct {
}

// RoleGetServerResponse is the response for the 'get' method.
type RoleGetServerResponse struct {
	status int
	err    *errors.Error
	body   *Role
}

// Body sets the value of the 'body' parameter.
//
//
func (r *RoleGetServerResponse) Body(value *Role) *RoleGetServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *RoleGetServerResponse) Status(value int) *RoleGetServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'get' method.
func (r *RoleGetServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// RoleUpdateServerRequest is the request for the 'update' method.
type RoleUpdateServerRequest struct {
	body *Role
}

// Body returns the value of the 'body' parameter.
//
//
func (r *RoleUpdateServerRequest) Body() *Role {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
//
func (r *RoleUpdateServerRequest) GetBody() (value *Role, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal request to the
// 'update' method.
func (r *RoleUpdateServerRequest) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(roleData)
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

// RoleUpdateServerResponse is the response for the 'update' method.
type RoleUpdateServerResponse struct {
	status int
	err    *errors.Error
	body   *Role
}

// Body sets the value of the 'body' parameter.
//
//
func (r *RoleUpdateServerResponse) Body(value *Role) *RoleUpdateServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *RoleUpdateServerResponse) Status(value int) *RoleUpdateServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'update' method.
func (r *RoleUpdateServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// dispatchRole navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchRole(w http.ResponseWriter, r *http.Request, server RoleServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "DELETE":
			adaptRoleDeleteRequest(w, r, server)
		case "GET":
			adaptRoleGetRequest(w, r, server)
		case "PATCH":
			adaptRoleUpdateRequest(w, r, server)
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

// readRoleDeleteRequest reads the given HTTP requests and translates it
// into an object of type RoleDeleteServerRequest.
func readRoleDeleteRequest(r *http.Request) (*RoleDeleteServerRequest, error) {
	var err error
	result := new(RoleDeleteServerRequest)
	return result, err
}

// writeRoleDeleteResponse translates the given request object into an
// HTTP response.
func writeRoleDeleteResponse(w http.ResponseWriter, r *RoleDeleteServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	return nil
}

// adaptRoleDeleteRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptRoleDeleteRequest(w http.ResponseWriter, r *http.Request, server RoleServer) {
	request, err := readRoleDeleteRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(RoleDeleteServerResponse)
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
	err = writeRoleDeleteResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// readRoleGetRequest reads the given HTTP requests and translates it
// into an object of type RoleGetServerRequest.
func readRoleGetRequest(r *http.Request) (*RoleGetServerRequest, error) {
	var err error
	result := new(RoleGetServerRequest)
	return result, err
}

// writeRoleGetResponse translates the given request object into an
// HTTP response.
func writeRoleGetResponse(w http.ResponseWriter, r *RoleGetServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptRoleGetRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptRoleGetRequest(w http.ResponseWriter, r *http.Request, server RoleServer) {
	request, err := readRoleGetRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(RoleGetServerResponse)
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
	err = writeRoleGetResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// readRoleUpdateRequest reads the given HTTP requests and translates it
// into an object of type RoleUpdateServerRequest.
func readRoleUpdateRequest(r *http.Request) (*RoleUpdateServerRequest, error) {
	var err error
	result := new(RoleUpdateServerRequest)
	err = result.unmarshal(r.Body)
	if err != nil {
		return nil, err
	}
	return result, err
}

// writeRoleUpdateResponse translates the given request object into an
// HTTP response.
func writeRoleUpdateResponse(w http.ResponseWriter, r *RoleUpdateServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptRoleUpdateRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptRoleUpdateRequest(w http.ResponseWriter, r *http.Request, server RoleServer) {
	request, err := readRoleUpdateRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(RoleUpdateServerResponse)
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
	err = writeRoleUpdateResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
