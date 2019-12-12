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
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// AddOnInstallationsServer represents the interface the manages the 'add_on_installations' resource.
type AddOnInstallationsServer interface {

	// Add handles a request for the 'add' method.
	//
	// Create a new add-on installation and add it to the collection of add-on installations on the cluster.
	Add(ctx context.Context, request *AddOnInstallationsAddServerRequest, response *AddOnInstallationsAddServerResponse) error

	// List handles a request for the 'list' method.
	//
	// Retrieves the list of add-on installations.
	List(ctx context.Context, request *AddOnInstallationsListServerRequest, response *AddOnInstallationsListServerResponse) error

	// Addoninstallation returns the target 'add_on_installation' server for the given identifier.
	//
	// Returns a reference to the service that manages a specific add-on installation.
	Addoninstallation(id string) AddOnInstallationServer
}

// AddOnInstallationsAddServerRequest is the request for the 'add' method.
type AddOnInstallationsAddServerRequest struct {
	body *AddOnInstallation
}

// Body returns the value of the 'body' parameter.
//
// Description of the add-on installation.
func (r *AddOnInstallationsAddServerRequest) Body() *AddOnInstallation {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
// Description of the add-on installation.
func (r *AddOnInstallationsAddServerRequest) GetBody() (value *AddOnInstallation, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal request to the
// 'add' method.
func (r *AddOnInstallationsAddServerRequest) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(addOnInstallationData)
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

// AddOnInstallationsAddServerResponse is the response for the 'add' method.
type AddOnInstallationsAddServerResponse struct {
	status int
	err    *errors.Error
	body   *AddOnInstallation
}

// Body sets the value of the 'body' parameter.
//
// Description of the add-on installation.
func (r *AddOnInstallationsAddServerResponse) Body(value *AddOnInstallation) *AddOnInstallationsAddServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *AddOnInstallationsAddServerResponse) Status(value int) *AddOnInstallationsAddServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'add' method.
func (r *AddOnInstallationsAddServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// AddOnInstallationsListServerRequest is the request for the 'list' method.
type AddOnInstallationsListServerRequest struct {
	order  *string
	page   *int
	search *string
	size   *int
}

// Order returns the value of the 'order' parameter.
//
// Order criteria.
//
// The syntax of this parameter is similar to the syntax of the _order by_ clause of
// a SQL statement, but using the names of the attributes of the add-on installation
// instead of the names of the columns of a table. For example, in order to sort the
// add-on installations descending by name the value should be:
//
// [source,sql]
// ----
// name desc
// ----
//
// If the parameter isn't provided, or if the value is empty, then the order of the
// results is undefined.
func (r *AddOnInstallationsListServerRequest) Order() string {
	if r != nil && r.order != nil {
		return *r.order
	}
	return ""
}

// GetOrder returns the value of the 'order' parameter and
// a flag indicating if the parameter has a value.
//
// Order criteria.
//
// The syntax of this parameter is similar to the syntax of the _order by_ clause of
// a SQL statement, but using the names of the attributes of the add-on installation
// instead of the names of the columns of a table. For example, in order to sort the
// add-on installations descending by name the value should be:
//
// [source,sql]
// ----
// name desc
// ----
//
// If the parameter isn't provided, or if the value is empty, then the order of the
// results is undefined.
func (r *AddOnInstallationsListServerRequest) GetOrder() (value string, ok bool) {
	ok = r != nil && r.order != nil
	if ok {
		value = *r.order
	}
	return
}

// Page returns the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
func (r *AddOnInstallationsListServerRequest) Page() int {
	if r != nil && r.page != nil {
		return *r.page
	}
	return 0
}

// GetPage returns the value of the 'page' parameter and
// a flag indicating if the parameter has a value.
//
// Index of the requested page, where one corresponds to the first page.
func (r *AddOnInstallationsListServerRequest) GetPage() (value int, ok bool) {
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
// The syntax of this parameter is similar to the syntax of the _where_ clause of an
// SQL statement, but using the names of the attributes of the add-on installation
// instead of the names of the columns of a table. For example, in order to retrieve
// all the add-on installations with a name starting with `my` the value should be:
//
// [source,sql]
// ----
// name like 'my%'
// ----
//
// If the parameter isn't provided, or if the value is empty, then all the add-on
// installations that the user has permission to see will be returned.
func (r *AddOnInstallationsListServerRequest) Search() string {
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
// The syntax of this parameter is similar to the syntax of the _where_ clause of an
// SQL statement, but using the names of the attributes of the add-on installation
// instead of the names of the columns of a table. For example, in order to retrieve
// all the add-on installations with a name starting with `my` the value should be:
//
// [source,sql]
// ----
// name like 'my%'
// ----
//
// If the parameter isn't provided, or if the value is empty, then all the add-on
// installations that the user has permission to see will be returned.
func (r *AddOnInstallationsListServerRequest) GetSearch() (value string, ok bool) {
	ok = r != nil && r.search != nil
	if ok {
		value = *r.search
	}
	return
}

// Size returns the value of the 'size' parameter.
//
// Maximum number of items that will be contained in the returned page.
func (r *AddOnInstallationsListServerRequest) Size() int {
	if r != nil && r.size != nil {
		return *r.size
	}
	return 0
}

// GetSize returns the value of the 'size' parameter and
// a flag indicating if the parameter has a value.
//
// Maximum number of items that will be contained in the returned page.
func (r *AddOnInstallationsListServerRequest) GetSize() (value int, ok bool) {
	ok = r != nil && r.size != nil
	if ok {
		value = *r.size
	}
	return
}

// AddOnInstallationsListServerResponse is the response for the 'list' method.
type AddOnInstallationsListServerResponse struct {
	status int
	err    *errors.Error
	items  *AddOnInstallationList
	page   *int
	size   *int
	total  *int
}

// Items sets the value of the 'items' parameter.
//
// Retrieved list of add-on installations.
func (r *AddOnInstallationsListServerResponse) Items(value *AddOnInstallationList) *AddOnInstallationsListServerResponse {
	r.items = value
	return r
}

// Page sets the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
func (r *AddOnInstallationsListServerResponse) Page(value int) *AddOnInstallationsListServerResponse {
	r.page = &value
	return r
}

// Size sets the value of the 'size' parameter.
//
// Maximum number of items that will be contained in the returned page.
func (r *AddOnInstallationsListServerResponse) Size(value int) *AddOnInstallationsListServerResponse {
	r.size = &value
	return r
}

// Total sets the value of the 'total' parameter.
//
// Total number of items of the collection that match the search criteria,
// regardless of the size of the page.
func (r *AddOnInstallationsListServerResponse) Total(value int) *AddOnInstallationsListServerResponse {
	r.total = &value
	return r
}

// Status sets the status code.
func (r *AddOnInstallationsListServerResponse) Status(value int) *AddOnInstallationsListServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'list' method.
func (r *AddOnInstallationsListServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data := new(addOnInstallationsListServerResponseData)
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

// addOnInstallationsListServerResponseData is the structure used internally to write the request of the
// 'list' method.
type addOnInstallationsListServerResponseData struct {
	Items addOnInstallationListData "json:\"items,omitempty\""
	Page  *int                      "json:\"page,omitempty\""
	Size  *int                      "json:\"size,omitempty\""
	Total *int                      "json:\"total,omitempty\""
}

// dispatchAddOnInstallations navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchAddOnInstallations(w http.ResponseWriter, r *http.Request, server AddOnInstallationsServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "POST":
			adaptAddOnInstallationsAddRequest(w, r, server)
		case "GET":
			adaptAddOnInstallationsListRequest(w, r, server)
		default:
			errors.SendMethodNotAllowed(w, r)
			return
		}
	} else {
		switch segments[0] {
		default:
			target := server.Addoninstallation(segments[0])
			if target == nil {
				errors.SendNotFound(w, r)
				return
			}
			dispatchAddOnInstallation(w, r, target, segments[1:])
		}
	}
}

// readAddOnInstallationsAddRequest reads the given HTTP requests and translates it
// into an object of type AddOnInstallationsAddServerRequest.
func readAddOnInstallationsAddRequest(r *http.Request) (*AddOnInstallationsAddServerRequest, error) {
	var err error
	result := new(AddOnInstallationsAddServerRequest)
	err = result.unmarshal(r.Body)
	if err != nil {
		return nil, err
	}
	return result, err
}

// writeAddOnInstallationsAddResponse translates the given request object into an
// HTTP response.
func writeAddOnInstallationsAddResponse(w http.ResponseWriter, r *AddOnInstallationsAddServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptAddOnInstallationsAddRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptAddOnInstallationsAddRequest(w http.ResponseWriter, r *http.Request, server AddOnInstallationsServer) {
	request, err := readAddOnInstallationsAddRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(AddOnInstallationsAddServerResponse)
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
	err = writeAddOnInstallationsAddResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// readAddOnInstallationsListRequest reads the given HTTP requests and translates it
// into an object of type AddOnInstallationsListServerRequest.
func readAddOnInstallationsListRequest(r *http.Request) (*AddOnInstallationsListServerRequest, error) {
	var err error
	result := new(AddOnInstallationsListServerRequest)
	query := r.URL.Query()
	result.order, err = helpers.ParseString(query, "order")
	if err != nil {
		return nil, err
	}
	result.page, err = helpers.ParseInteger(query, "page")
	if err != nil {
		return nil, err
	}
	if result.page == nil {
		result.page = helpers.NewInteger(1)
	}
	result.search, err = helpers.ParseString(query, "search")
	if err != nil {
		return nil, err
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

// writeAddOnInstallationsListResponse translates the given request object into an
// HTTP response.
func writeAddOnInstallationsListResponse(w http.ResponseWriter, r *AddOnInstallationsListServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptAddOnInstallationsListRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptAddOnInstallationsListRequest(w http.ResponseWriter, r *http.Request, server AddOnInstallationsServer) {
	request, err := readAddOnInstallationsListRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(AddOnInstallationsListServerResponse)
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
	err = writeAddOnInstallationsListResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
