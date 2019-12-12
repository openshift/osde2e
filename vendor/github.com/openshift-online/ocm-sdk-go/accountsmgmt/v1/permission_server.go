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

// PermissionServer represents the interface the manages the 'permission' resource.
type PermissionServer interface {

	// Delete handles a request for the 'delete' method.
	//
	// Deletes the permission.
	Delete(ctx context.Context, request *PermissionDeleteServerRequest, response *PermissionDeleteServerResponse) error

	// Get handles a request for the 'get' method.
	//
	// Retrieves the details of the permission.
	Get(ctx context.Context, request *PermissionGetServerRequest, response *PermissionGetServerResponse) error
}

// PermissionDeleteServerRequest is the request for the 'delete' method.
type PermissionDeleteServerRequest struct {
}

// PermissionDeleteServerResponse is the response for the 'delete' method.
type PermissionDeleteServerResponse struct {
	status int
	err    *errors.Error
}

// Status sets the status code.
func (r *PermissionDeleteServerResponse) Status(value int) *PermissionDeleteServerResponse {
	r.status = value
	return r
}

// PermissionGetServerRequest is the request for the 'get' method.
type PermissionGetServerRequest struct {
}

// PermissionGetServerResponse is the response for the 'get' method.
type PermissionGetServerResponse struct {
	status int
	err    *errors.Error
	body   *Permission
}

// Body sets the value of the 'body' parameter.
//
//
func (r *PermissionGetServerResponse) Body(value *Permission) *PermissionGetServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *PermissionGetServerResponse) Status(value int) *PermissionGetServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'get' method.
func (r *PermissionGetServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// dispatchPermission navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchPermission(w http.ResponseWriter, r *http.Request, server PermissionServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "DELETE":
			adaptPermissionDeleteRequest(w, r, server)
		case "GET":
			adaptPermissionGetRequest(w, r, server)
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

// readPermissionDeleteRequest reads the given HTTP requests and translates it
// into an object of type PermissionDeleteServerRequest.
func readPermissionDeleteRequest(r *http.Request) (*PermissionDeleteServerRequest, error) {
	var err error
	result := new(PermissionDeleteServerRequest)
	return result, err
}

// writePermissionDeleteResponse translates the given request object into an
// HTTP response.
func writePermissionDeleteResponse(w http.ResponseWriter, r *PermissionDeleteServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	return nil
}

// adaptPermissionDeleteRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptPermissionDeleteRequest(w http.ResponseWriter, r *http.Request, server PermissionServer) {
	request, err := readPermissionDeleteRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(PermissionDeleteServerResponse)
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
	err = writePermissionDeleteResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// readPermissionGetRequest reads the given HTTP requests and translates it
// into an object of type PermissionGetServerRequest.
func readPermissionGetRequest(r *http.Request) (*PermissionGetServerRequest, error) {
	var err error
	result := new(PermissionGetServerRequest)
	return result, err
}

// writePermissionGetResponse translates the given request object into an
// HTTP response.
func writePermissionGetResponse(w http.ResponseWriter, r *PermissionGetServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptPermissionGetRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptPermissionGetRequest(w http.ResponseWriter, r *http.Request, server PermissionServer) {
	request, err := readPermissionGetRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(PermissionGetServerResponse)
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
	err = writePermissionGetResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
