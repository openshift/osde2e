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

// AccountServer represents the interface the manages the 'account' resource.
type AccountServer interface {

	// Get handles a request for the 'get' method.
	//
	// Retrieves the details of the account.
	Get(ctx context.Context, request *AccountGetServerRequest, response *AccountGetServerResponse) error

	// Update handles a request for the 'update' method.
	//
	// Updates the account.
	Update(ctx context.Context, request *AccountUpdateServerRequest, response *AccountUpdateServerResponse) error
}

// AccountGetServerRequest is the request for the 'get' method.
type AccountGetServerRequest struct {
}

// AccountGetServerResponse is the response for the 'get' method.
type AccountGetServerResponse struct {
	status int
	err    *errors.Error
	body   *Account
}

// Body sets the value of the 'body' parameter.
//
//
func (r *AccountGetServerResponse) Body(value *Account) *AccountGetServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *AccountGetServerResponse) Status(value int) *AccountGetServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'get' method.
func (r *AccountGetServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// AccountUpdateServerRequest is the request for the 'update' method.
type AccountUpdateServerRequest struct {
	body *Account
}

// Body returns the value of the 'body' parameter.
//
//
func (r *AccountUpdateServerRequest) Body() *Account {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
//
func (r *AccountUpdateServerRequest) GetBody() (value *Account, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal request to the
// 'update' method.
func (r *AccountUpdateServerRequest) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(accountData)
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

// AccountUpdateServerResponse is the response for the 'update' method.
type AccountUpdateServerResponse struct {
	status int
	err    *errors.Error
	body   *Account
}

// Body sets the value of the 'body' parameter.
//
//
func (r *AccountUpdateServerResponse) Body(value *Account) *AccountUpdateServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *AccountUpdateServerResponse) Status(value int) *AccountUpdateServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'update' method.
func (r *AccountUpdateServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// dispatchAccount navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchAccount(w http.ResponseWriter, r *http.Request, server AccountServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "GET":
			adaptAccountGetRequest(w, r, server)
		case "PATCH":
			adaptAccountUpdateRequest(w, r, server)
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

// readAccountGetRequest reads the given HTTP requests and translates it
// into an object of type AccountGetServerRequest.
func readAccountGetRequest(r *http.Request) (*AccountGetServerRequest, error) {
	var err error
	result := new(AccountGetServerRequest)
	return result, err
}

// writeAccountGetResponse translates the given request object into an
// HTTP response.
func writeAccountGetResponse(w http.ResponseWriter, r *AccountGetServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptAccountGetRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptAccountGetRequest(w http.ResponseWriter, r *http.Request, server AccountServer) {
	request, err := readAccountGetRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(AccountGetServerResponse)
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
	err = writeAccountGetResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// readAccountUpdateRequest reads the given HTTP requests and translates it
// into an object of type AccountUpdateServerRequest.
func readAccountUpdateRequest(r *http.Request) (*AccountUpdateServerRequest, error) {
	var err error
	result := new(AccountUpdateServerRequest)
	err = result.unmarshal(r.Body)
	if err != nil {
		return nil, err
	}
	return result, err
}

// writeAccountUpdateResponse translates the given request object into an
// HTTP response.
func writeAccountUpdateResponse(w http.ResponseWriter, r *AccountUpdateServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptAccountUpdateRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptAccountUpdateRequest(w http.ResponseWriter, r *http.Request, server AccountServer) {
	request, err := readAccountUpdateRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(AccountUpdateServerResponse)
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
	err = writeAccountUpdateResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
