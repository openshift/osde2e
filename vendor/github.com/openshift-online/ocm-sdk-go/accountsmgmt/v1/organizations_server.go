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

// OrganizationsServer represents the interface the manages the 'organizations' resource.
type OrganizationsServer interface {

	// Add handles a request for the 'add' method.
	//
	// Creates a new organization.
	Add(ctx context.Context, request *OrganizationsAddServerRequest, response *OrganizationsAddServerResponse) error

	// List handles a request for the 'list' method.
	//
	// Retrieves a list of organizations.
	List(ctx context.Context, request *OrganizationsListServerRequest, response *OrganizationsListServerResponse) error

	// Organization returns the target 'organization' server for the given identifier.
	//
	// Reference to the service that manages a specific organization.
	Organization(id string) OrganizationServer
}

// OrganizationsAddServerRequest is the request for the 'add' method.
type OrganizationsAddServerRequest struct {
	body *Organization
}

// Body returns the value of the 'body' parameter.
//
// Organization data.
func (r *OrganizationsAddServerRequest) Body() *Organization {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
// Organization data.
func (r *OrganizationsAddServerRequest) GetBody() (value *Organization, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal request to the
// 'add' method.
func (r *OrganizationsAddServerRequest) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(organizationData)
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

// OrganizationsAddServerResponse is the response for the 'add' method.
type OrganizationsAddServerResponse struct {
	status int
	err    *errors.Error
	body   *Organization
}

// Body sets the value of the 'body' parameter.
//
// Organization data.
func (r *OrganizationsAddServerResponse) Body(value *Organization) *OrganizationsAddServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *OrganizationsAddServerResponse) Status(value int) *OrganizationsAddServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'add' method.
func (r *OrganizationsAddServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// OrganizationsListServerRequest is the request for the 'list' method.
type OrganizationsListServerRequest struct {
	page   *int
	search *string
	size   *int
	total  *int
}

// Page returns the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
//
// Default value is `1`.
func (r *OrganizationsListServerRequest) Page() int {
	if r != nil && r.page != nil {
		return *r.page
	}
	return 0
}

// GetPage returns the value of the 'page' parameter and
// a flag indicating if the parameter has a value.
//
// Index of the requested page, where one corresponds to the first page.
//
// Default value is `1`.
func (r *OrganizationsListServerRequest) GetPage() (value int, ok bool) {
	ok = r != nil && r.page != nil
	if ok {
		value = *r.page
	}
	return
}

// Search returns the value of the 'search' parameter.
//
// Search criteria.
//
// The syntax of this parameter is similar to the syntax of the _where_ clause
// of an SQL statement, but using the names of the attributes of the organization
// instead of the names of the columns of a table. For example, in order to
// retrieve organizations with name starting with my:
//
// [source,sql]
// ----
// name like 'my%'
// ----
//
// If the parameter isn't provided, or if the value is empty, then all the
// items that the user has permission to see will be returned.
func (r *OrganizationsListServerRequest) Search() string {
	if r != nil && r.search != nil {
		return *r.search
	}
	return ""
}

// GetSearch returns the value of the 'search' parameter and
// a flag indicating if the parameter has a value.
//
// Search criteria.
//
// The syntax of this parameter is similar to the syntax of the _where_ clause
// of an SQL statement, but using the names of the attributes of the organization
// instead of the names of the columns of a table. For example, in order to
// retrieve organizations with name starting with my:
//
// [source,sql]
// ----
// name like 'my%'
// ----
//
// If the parameter isn't provided, or if the value is empty, then all the
// items that the user has permission to see will be returned.
func (r *OrganizationsListServerRequest) GetSearch() (value string, ok bool) {
	ok = r != nil && r.search != nil
	if ok {
		value = *r.search
	}
	return
}

// Size returns the value of the 'size' parameter.
//
// Maximum number of items that will be contained in the returned page.
//
// Default value is `100`.
func (r *OrganizationsListServerRequest) Size() int {
	if r != nil && r.size != nil {
		return *r.size
	}
	return 0
}

// GetSize returns the value of the 'size' parameter and
// a flag indicating if the parameter has a value.
//
// Maximum number of items that will be contained in the returned page.
//
// Default value is `100`.
func (r *OrganizationsListServerRequest) GetSize() (value int, ok bool) {
	ok = r != nil && r.size != nil
	if ok {
		value = *r.size
	}
	return
}

// Total returns the value of the 'total' parameter.
//
// Total number of items of the collection that match the search criteria,
// regardless of the size of the page.
func (r *OrganizationsListServerRequest) Total() int {
	if r != nil && r.total != nil {
		return *r.total
	}
	return 0
}

// GetTotal returns the value of the 'total' parameter and
// a flag indicating if the parameter has a value.
//
// Total number of items of the collection that match the search criteria,
// regardless of the size of the page.
func (r *OrganizationsListServerRequest) GetTotal() (value int, ok bool) {
	ok = r != nil && r.total != nil
	if ok {
		value = *r.total
	}
	return
}

// OrganizationsListServerResponse is the response for the 'list' method.
type OrganizationsListServerResponse struct {
	status int
	err    *errors.Error
	items  *OrganizationList
	page   *int
	size   *int
	total  *int
}

// Items sets the value of the 'items' parameter.
//
// Retrieved list of organizations.
func (r *OrganizationsListServerResponse) Items(value *OrganizationList) *OrganizationsListServerResponse {
	r.items = value
	return r
}

// Page sets the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
//
// Default value is `1`.
func (r *OrganizationsListServerResponse) Page(value int) *OrganizationsListServerResponse {
	r.page = &value
	return r
}

// Size sets the value of the 'size' parameter.
//
// Maximum number of items that will be contained in the returned page.
//
// Default value is `100`.
func (r *OrganizationsListServerResponse) Size(value int) *OrganizationsListServerResponse {
	r.size = &value
	return r
}

// Total sets the value of the 'total' parameter.
//
// Total number of items of the collection that match the search criteria,
// regardless of the size of the page.
func (r *OrganizationsListServerResponse) Total(value int) *OrganizationsListServerResponse {
	r.total = &value
	return r
}

// Status sets the status code.
func (r *OrganizationsListServerResponse) Status(value int) *OrganizationsListServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'list' method.
func (r *OrganizationsListServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data := new(organizationsListServerResponseData)
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

// organizationsListServerResponseData is the structure used internally to write the request of the
// 'list' method.
type organizationsListServerResponseData struct {
	Items organizationListData "json:\"items,omitempty\""
	Page  *int                 "json:\"page,omitempty\""
	Size  *int                 "json:\"size,omitempty\""
	Total *int                 "json:\"total,omitempty\""
}

// dispatchOrganizations navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchOrganizations(w http.ResponseWriter, r *http.Request, server OrganizationsServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "POST":
			adaptOrganizationsAddRequest(w, r, server)
		case "GET":
			adaptOrganizationsListRequest(w, r, server)
		default:
			errors.SendMethodNotAllowed(w, r)
			return
		}
	} else {
		switch segments[0] {
		default:
			target := server.Organization(segments[0])
			if target == nil {
				errors.SendNotFound(w, r)
				return
			}
			dispatchOrganization(w, r, target, segments[1:])
		}
	}
}

// readOrganizationsAddRequest reads the given HTTP requests and translates it
// into an object of type OrganizationsAddServerRequest.
func readOrganizationsAddRequest(r *http.Request) (*OrganizationsAddServerRequest, error) {
	var err error
	result := new(OrganizationsAddServerRequest)
	err = result.unmarshal(r.Body)
	if err != nil {
		return nil, err
	}
	return result, err
}

// writeOrganizationsAddResponse translates the given request object into an
// HTTP response.
func writeOrganizationsAddResponse(w http.ResponseWriter, r *OrganizationsAddServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptOrganizationsAddRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptOrganizationsAddRequest(w http.ResponseWriter, r *http.Request, server OrganizationsServer) {
	request, err := readOrganizationsAddRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(OrganizationsAddServerResponse)
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
	err = writeOrganizationsAddResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// readOrganizationsListRequest reads the given HTTP requests and translates it
// into an object of type OrganizationsListServerRequest.
func readOrganizationsListRequest(r *http.Request) (*OrganizationsListServerRequest, error) {
	var err error
	result := new(OrganizationsListServerRequest)
	query := r.URL.Query()
	result.page, err = helpers.ParseInteger(query, "page")
	if err != nil {
		return nil, err
	}
	result.search, err = helpers.ParseString(query, "search")
	if err != nil {
		return nil, err
	}
	result.size, err = helpers.ParseInteger(query, "size")
	if err != nil {
		return nil, err
	}
	result.total, err = helpers.ParseInteger(query, "total")
	if err != nil {
		return nil, err
	}
	return result, err
}

// writeOrganizationsListResponse translates the given request object into an
// HTTP response.
func writeOrganizationsListResponse(w http.ResponseWriter, r *OrganizationsListServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptOrganizationsListRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptOrganizationsListRequest(w http.ResponseWriter, r *http.Request, server OrganizationsServer) {
	request, err := readOrganizationsListRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(OrganizationsListServerResponse)
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
	err = writeOrganizationsListResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
