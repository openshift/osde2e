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
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// PermissionsServer represents the interface the manages the 'permissions' resource.
type PermissionsServer interface {

	// Add handles a request for the 'add' method.
	//
	// Creates a new permission.
	Add(ctx context.Context, request *PermissionsAddServerRequest, response *PermissionsAddServerResponse) error

	// List handles a request for the 'list' method.
	//
	// Retrieves a list of permissions.
	List(ctx context.Context, request *PermissionsListServerRequest, response *PermissionsListServerResponse) error

	// Permission returns the target 'permission' server for the given identifier.
	//
	// Reference to the service that manages an specific permission.
	Permission(id string) PermissionServer
}

// PermissionsAddServerRequest is the request for the 'add' method.
type PermissionsAddServerRequest struct {
	body *Permission
}

// Body returns the value of the 'body' parameter.
//
// Permission data.
func (r *PermissionsAddServerRequest) Body() *Permission {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
// Permission data.
func (r *PermissionsAddServerRequest) GetBody() (value *Permission, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal request to the
// 'add' method.
func (r *PermissionsAddServerRequest) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(permissionData)
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

// PermissionsAddServerResponse is the response for the 'add' method.
type PermissionsAddServerResponse struct {
	status int
	err    *errors.Error
	body   *Permission
}

// Body sets the value of the 'body' parameter.
//
// Permission data.
func (r *PermissionsAddServerResponse) Body(value *Permission) *PermissionsAddServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *PermissionsAddServerResponse) Status(value int) *PermissionsAddServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'add' method.
func (r *PermissionsAddServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// PermissionsListServerRequest is the request for the 'list' method.
type PermissionsListServerRequest struct {
	page *int
	size *int
}

// Page returns the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
func (r *PermissionsListServerRequest) Page() int {
	if r != nil && r.page != nil {
		return *r.page
	}
	return 0
}

// GetPage returns the value of the 'page' parameter and
// a flag indicating if the parameter has a value.
//
// Index of the requested page, where one corresponds to the first page.
func (r *PermissionsListServerRequest) GetPage() (value int, ok bool) {
	ok = r != nil && r.page != nil
	if ok {
		value = *r.page
	}
	return
}

// Size returns the value of the 'size' parameter.
//
// Maximum number of items that will be contained in the returned page.
func (r *PermissionsListServerRequest) Size() int {
	if r != nil && r.size != nil {
		return *r.size
	}
	return 0
}

// GetSize returns the value of the 'size' parameter and
// a flag indicating if the parameter has a value.
//
// Maximum number of items that will be contained in the returned page.
func (r *PermissionsListServerRequest) GetSize() (value int, ok bool) {
	ok = r != nil && r.size != nil
	if ok {
		value = *r.size
	}
	return
}

// PermissionsListServerResponse is the response for the 'list' method.
type PermissionsListServerResponse struct {
	status int
	err    *errors.Error
	items  *PermissionList
	page   *int
	size   *int
	total  *int
}

// Items sets the value of the 'items' parameter.
//
// Retrieved list of permissions.
func (r *PermissionsListServerResponse) Items(value *PermissionList) *PermissionsListServerResponse {
	r.items = value
	return r
}

// Page sets the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
func (r *PermissionsListServerResponse) Page(value int) *PermissionsListServerResponse {
	r.page = &value
	return r
}

// Size sets the value of the 'size' parameter.
//
// Maximum number of items that will be contained in the returned page.
func (r *PermissionsListServerResponse) Size(value int) *PermissionsListServerResponse {
	r.size = &value
	return r
}

// Total sets the value of the 'total' parameter.
//
// Total number of items of the collection that match the search criteria,
// regardless of the size of the page.
func (r *PermissionsListServerResponse) Total(value int) *PermissionsListServerResponse {
	r.total = &value
	return r
}

// Status sets the status code.
func (r *PermissionsListServerResponse) Status(value int) *PermissionsListServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'list' method.
func (r *PermissionsListServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data := new(permissionsListServerResponseData)
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

// permissionsListServerResponseData is the structure used internally to write the request of the
// 'list' method.
type permissionsListServerResponseData struct {
	Items permissionListData "json:\"items,omitempty\""
	Page  *int               "json:\"page,omitempty\""
	Size  *int               "json:\"size,omitempty\""
	Total *int               "json:\"total,omitempty\""
}

// dispatchPermissions navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchPermissions(w http.ResponseWriter, r *http.Request, server PermissionsServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "POST":
			adaptPermissionsAddRequest(w, r, server)
		case "GET":
			adaptPermissionsListRequest(w, r, server)
		default:
			errors.SendMethodNotAllowed(w, r)
			return
		}
	} else {
		switch segments[0] {
		default:
			target := server.Permission(segments[0])
			if target == nil {
				errors.SendNotFound(w, r)
				return
			}
			dispatchPermission(w, r, target, segments[1:])
		}
	}
}

// readPermissionsAddRequest reads the given HTTP requests and translates it
// into an object of type PermissionsAddServerRequest.
func readPermissionsAddRequest(r *http.Request) (*PermissionsAddServerRequest, error) {
	var err error
	result := new(PermissionsAddServerRequest)
	err = result.unmarshal(r.Body)
	if err != nil {
		return nil, err
	}
	return result, err
}

// writePermissionsAddResponse translates the given request object into an
// HTTP response.
func writePermissionsAddResponse(w http.ResponseWriter, r *PermissionsAddServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptPermissionsAddRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptPermissionsAddRequest(w http.ResponseWriter, r *http.Request, server PermissionsServer) {
	request, err := readPermissionsAddRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(PermissionsAddServerResponse)
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
	err = writePermissionsAddResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// readPermissionsListRequest reads the given HTTP requests and translates it
// into an object of type PermissionsListServerRequest.
func readPermissionsListRequest(r *http.Request) (*PermissionsListServerRequest, error) {
	var err error
	result := new(PermissionsListServerRequest)
	query := r.URL.Query()
	result.page, err = helpers.ParseInteger(query, "page")
	if err != nil {
		return nil, err
	}
	if result.page == nil {
		result.page = helpers.NewInteger(1)
	}
	result.size, err = helpers.ParseInteger(query, "size")
	if err != nil {
		return nil, err
	}
	if result.size == nil {
		result.size = helpers.NewInteger(100)
	}
	return result, err
}

// writePermissionsListResponse translates the given request object into an
// HTTP response.
func writePermissionsListResponse(w http.ResponseWriter, r *PermissionsListServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptPermissionsListRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptPermissionsListRequest(w http.ResponseWriter, r *http.Request, server PermissionsServer) {
	request, err := readPermissionsListRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(PermissionsListServerResponse)
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
	err = writePermissionsListResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
