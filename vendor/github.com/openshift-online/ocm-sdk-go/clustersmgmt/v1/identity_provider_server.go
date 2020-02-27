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

// IdentityProviderServer represents the interface the manages the 'identity_provider' resource.
type IdentityProviderServer interface {

	// Delete handles a request for the 'delete' method.
	//
	// Deletes the identity provider.
	Delete(ctx context.Context, request *IdentityProviderDeleteServerRequest, response *IdentityProviderDeleteServerResponse) error

	// Get handles a request for the 'get' method.
	//
	// Retrieves the details of the identity provider.
	Get(ctx context.Context, request *IdentityProviderGetServerRequest, response *IdentityProviderGetServerResponse) error
}

// IdentityProviderDeleteServerRequest is the request for the 'delete' method.
type IdentityProviderDeleteServerRequest struct {
}

// IdentityProviderDeleteServerResponse is the response for the 'delete' method.
type IdentityProviderDeleteServerResponse struct {
	status int
	err    *errors.Error
}

// Status sets the status code.
func (r *IdentityProviderDeleteServerResponse) Status(value int) *IdentityProviderDeleteServerResponse {
	r.status = value
	return r
}

// IdentityProviderGetServerRequest is the request for the 'get' method.
type IdentityProviderGetServerRequest struct {
}

// IdentityProviderGetServerResponse is the response for the 'get' method.
type IdentityProviderGetServerResponse struct {
	status int
	err    *errors.Error
	body   *IdentityProvider
}

// Body sets the value of the 'body' parameter.
//
//
func (r *IdentityProviderGetServerResponse) Body(value *IdentityProvider) *IdentityProviderGetServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *IdentityProviderGetServerResponse) Status(value int) *IdentityProviderGetServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'get' method.
func (r *IdentityProviderGetServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// dispatchIdentityProvider navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchIdentityProvider(w http.ResponseWriter, r *http.Request, server IdentityProviderServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "DELETE":
			adaptIdentityProviderDeleteRequest(w, r, server)
		case "GET":
			adaptIdentityProviderGetRequest(w, r, server)
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

// readIdentityProviderDeleteRequest reads the given HTTP requests and translates it
// into an object of type IdentityProviderDeleteServerRequest.
func readIdentityProviderDeleteRequest(r *http.Request) (*IdentityProviderDeleteServerRequest, error) {
	var err error
	result := new(IdentityProviderDeleteServerRequest)
	return result, err
}

// writeIdentityProviderDeleteResponse translates the given request object into an
// HTTP response.
func writeIdentityProviderDeleteResponse(w http.ResponseWriter, r *IdentityProviderDeleteServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	return nil
}

// adaptIdentityProviderDeleteRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptIdentityProviderDeleteRequest(w http.ResponseWriter, r *http.Request, server IdentityProviderServer) {
	request, err := readIdentityProviderDeleteRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(IdentityProviderDeleteServerResponse)
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
	err = writeIdentityProviderDeleteResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// readIdentityProviderGetRequest reads the given HTTP requests and translates it
// into an object of type IdentityProviderGetServerRequest.
func readIdentityProviderGetRequest(r *http.Request) (*IdentityProviderGetServerRequest, error) {
	var err error
	result := new(IdentityProviderGetServerRequest)
	return result, err
}

// writeIdentityProviderGetResponse translates the given request object into an
// HTTP response.
func writeIdentityProviderGetResponse(w http.ResponseWriter, r *IdentityProviderGetServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptIdentityProviderGetRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptIdentityProviderGetRequest(w http.ResponseWriter, r *http.Request, server IdentityProviderServer) {
	request, err := readIdentityProviderGetRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(IdentityProviderGetServerResponse)
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
	err = writeIdentityProviderGetResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
