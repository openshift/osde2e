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

// UsersServer represents the interface the manages the 'users' resource.
type UsersServer interface {

	// Add handles a request for the 'add' method.
	//
	// Adds a new user to the group.
	Add(ctx context.Context, request *UsersAddServerRequest, response *UsersAddServerResponse) error

	// List handles a request for the 'list' method.
	//
	// Retrieves the list of users.
	List(ctx context.Context, request *UsersListServerRequest, response *UsersListServerResponse) error

	// User returns the target 'user' server for the given identifier.
	//
	// Reference to the service that manages an specific user.
	User(id string) UserServer
}

// UsersAddServerRequest is the request for the 'add' method.
type UsersAddServerRequest struct {
	body *User
}

// Body returns the value of the 'body' parameter.
//
// Description of the user.
func (r *UsersAddServerRequest) Body() *User {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
// Description of the user.
func (r *UsersAddServerRequest) GetBody() (value *User, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal request to the
// 'add' method.
func (r *UsersAddServerRequest) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(userData)
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

// UsersAddServerResponse is the response for the 'add' method.
type UsersAddServerResponse struct {
	status int
	err    *errors.Error
	body   *User
}

// Body sets the value of the 'body' parameter.
//
// Description of the user.
func (r *UsersAddServerResponse) Body(value *User) *UsersAddServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *UsersAddServerResponse) Status(value int) *UsersAddServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'add' method.
func (r *UsersAddServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// UsersListServerRequest is the request for the 'list' method.
type UsersListServerRequest struct {
}

// UsersListServerResponse is the response for the 'list' method.
type UsersListServerResponse struct {
	status int
	err    *errors.Error
	items  *UserList
	page   *int
	size   *int
	total  *int
}

// Items sets the value of the 'items' parameter.
//
// Retrieved list of users.
func (r *UsersListServerResponse) Items(value *UserList) *UsersListServerResponse {
	r.items = value
	return r
}

// Page sets the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
func (r *UsersListServerResponse) Page(value int) *UsersListServerResponse {
	r.page = &value
	return r
}

// Size sets the value of the 'size' parameter.
//
// Number of items contained in the returned page.
func (r *UsersListServerResponse) Size(value int) *UsersListServerResponse {
	r.size = &value
	return r
}

// Total sets the value of the 'total' parameter.
//
// Total number of items of the collection.
func (r *UsersListServerResponse) Total(value int) *UsersListServerResponse {
	r.total = &value
	return r
}

// Status sets the status code.
func (r *UsersListServerResponse) Status(value int) *UsersListServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'list' method.
func (r *UsersListServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data := new(usersListServerResponseData)
	data.Items, err = r.items.wrap()
	if err != nil {
		return err
	}
	data.Page = r.page
	data.Size = r.size
	data.Total = r.total
	err = encoder.Encode(data)
	return err
}

// usersListServerResponseData is the structure used internally to write the request of the
// 'list' method.
type usersListServerResponseData struct {
	Items userListData "json:\"items,omitempty\""
	Page  *int         "json:\"page,omitempty\""
	Size  *int         "json:\"size,omitempty\""
	Total *int         "json:\"total,omitempty\""
}

// dispatchUsers navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchUsers(w http.ResponseWriter, r *http.Request, server UsersServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "POST":
			adaptUsersAddRequest(w, r, server)
		case "GET":
			adaptUsersListRequest(w, r, server)
		default:
			errors.SendMethodNotAllowed(w, r)
			return
		}
	} else {
		switch segments[0] {
		default:
			target := server.User(segments[0])
			if target == nil {
				errors.SendNotFound(w, r)
				return
			}
			dispatchUser(w, r, target, segments[1:])
		}
	}
}

// readUsersAddRequest reads the given HTTP requests and translates it
// into an object of type UsersAddServerRequest.
func readUsersAddRequest(r *http.Request) (*UsersAddServerRequest, error) {
	var err error
	result := new(UsersAddServerRequest)
	err = result.unmarshal(r.Body)
	if err != nil {
		return nil, err
	}
	return result, err
}

// writeUsersAddResponse translates the given request object into an
// HTTP response.
func writeUsersAddResponse(w http.ResponseWriter, r *UsersAddServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptUsersAddRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptUsersAddRequest(w http.ResponseWriter, r *http.Request, server UsersServer) {
	request, err := readUsersAddRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(UsersAddServerResponse)
	response.status = 201
	err = server.Add(r.Context(), request, response)
	if err != nil {
		glog.Errorf(
			"Can't process request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	err = writeUsersAddResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// readUsersListRequest reads the given HTTP requests and translates it
// into an object of type UsersListServerRequest.
func readUsersListRequest(r *http.Request) (*UsersListServerRequest, error) {
	var err error
	result := new(UsersListServerRequest)
	return result, err
}

// writeUsersListResponse translates the given request object into an
// HTTP response.
func writeUsersListResponse(w http.ResponseWriter, r *UsersListServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptUsersListRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptUsersListRequest(w http.ResponseWriter, r *http.Request, server UsersServer) {
	request, err := readUsersListRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(UsersListServerResponse)
	response.status = 200
	err = server.List(r.Context(), request, response)
	if err != nil {
		glog.Errorf(
			"Can't process request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	err = writeUsersListResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
