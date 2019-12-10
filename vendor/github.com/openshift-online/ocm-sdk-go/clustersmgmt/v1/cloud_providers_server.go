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

// CloudProvidersServer represents the interface the manages the 'cloud_providers' resource.
type CloudProvidersServer interface {

	// List handles a request for the 'list' method.
	//
	// Retrieves the list of cloud providers.
	List(ctx context.Context, request *CloudProvidersListServerRequest, response *CloudProvidersListServerResponse) error

	// CloudProvider returns the target 'cloud_provider' server for the given identifier.
	//
	// Returns a reference to the service that manages an specific cloud provider.
	CloudProvider(id string) CloudProviderServer
}

// CloudProvidersListServerRequest is the request for the 'list' method.
type CloudProvidersListServerRequest struct {
	order  *string
	page   *int
	search *string
	size   *int
	total  *int
}

// Order returns the value of the 'order' parameter.
//
// Order criteria.
//
// The syntax of this parameter is similar to the syntax of the _order by_ clause of
// a SQL statement, but using the names of the attributes of the cloud provider
// instead of the names of the columns of a table. For example, in order to sort the
// clusters descending by name identifier the value should be:
//
// [source,sql]
// ----
// name desc
// ----
//
// If the parameter isn't provided, or if the value is empty, then the order of the
// results is undefined.
func (r *CloudProvidersListServerRequest) Order() string {
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
// a SQL statement, but using the names of the attributes of the cloud provider
// instead of the names of the columns of a table. For example, in order to sort the
// clusters descending by name identifier the value should be:
//
// [source,sql]
// ----
// name desc
// ----
//
// If the parameter isn't provided, or if the value is empty, then the order of the
// results is undefined.
func (r *CloudProvidersListServerRequest) GetOrder() (value string, ok bool) {
	ok = r != nil && r.order != nil
	if ok {
		value = *r.order
	}
	return
}

// Page returns the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
//
// Default value is `1`.
func (r *CloudProvidersListServerRequest) Page() int {
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
func (r *CloudProvidersListServerRequest) GetPage() (value int, ok bool) {
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
// The syntax of this parameter is similar to the syntax of the _where_ clause of a
// SQL statement, but using the names of the attributes of the cloud provider
// instead of the names of the columns of a table. For example, in order to retrieve
// all the cloud providers with a name starting with `A` the value should be:
//
// [source,sql]
// ----
// name like 'A%'
// ----
//
// If the parameter isn't provided, or if the value is empty, then all the clusters
// that the user has permission to see will be returned.
func (r *CloudProvidersListServerRequest) Search() string {
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
// The syntax of this parameter is similar to the syntax of the _where_ clause of a
// SQL statement, but using the names of the attributes of the cloud provider
// instead of the names of the columns of a table. For example, in order to retrieve
// all the cloud providers with a name starting with `A` the value should be:
//
// [source,sql]
// ----
// name like 'A%'
// ----
//
// If the parameter isn't provided, or if the value is empty, then all the clusters
// that the user has permission to see will be returned.
func (r *CloudProvidersListServerRequest) GetSearch() (value string, ok bool) {
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
func (r *CloudProvidersListServerRequest) Size() int {
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
func (r *CloudProvidersListServerRequest) GetSize() (value int, ok bool) {
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
func (r *CloudProvidersListServerRequest) Total() int {
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
func (r *CloudProvidersListServerRequest) GetTotal() (value int, ok bool) {
	ok = r != nil && r.total != nil
	if ok {
		value = *r.total
	}
	return
}

// CloudProvidersListServerResponse is the response for the 'list' method.
type CloudProvidersListServerResponse struct {
	status int
	err    *errors.Error
	items  *CloudProviderList
	page   *int
	size   *int
	total  *int
}

// Items sets the value of the 'items' parameter.
//
// Retrieved list of cloud providers.
func (r *CloudProvidersListServerResponse) Items(value *CloudProviderList) *CloudProvidersListServerResponse {
	r.items = value
	return r
}

// Page sets the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
//
// Default value is `1`.
func (r *CloudProvidersListServerResponse) Page(value int) *CloudProvidersListServerResponse {
	r.page = &value
	return r
}

// Size sets the value of the 'size' parameter.
//
// Maximum number of items that will be contained in the returned page.
//
// Default value is `100`.
func (r *CloudProvidersListServerResponse) Size(value int) *CloudProvidersListServerResponse {
	r.size = &value
	return r
}

// Total sets the value of the 'total' parameter.
//
// Total number of items of the collection that match the search criteria,
// regardless of the size of the page.
func (r *CloudProvidersListServerResponse) Total(value int) *CloudProvidersListServerResponse {
	r.total = &value
	return r
}

// Status sets the status code.
func (r *CloudProvidersListServerResponse) Status(value int) *CloudProvidersListServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'list' method.
func (r *CloudProvidersListServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data := new(cloudProvidersListServerResponseData)
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

// cloudProvidersListServerResponseData is the structure used internally to write the request of the
// 'list' method.
type cloudProvidersListServerResponseData struct {
	Items cloudProviderListData "json:\"items,omitempty\""
	Page  *int                  "json:\"page,omitempty\""
	Size  *int                  "json:\"size,omitempty\""
	Total *int                  "json:\"total,omitempty\""
}

// dispatchCloudProviders navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchCloudProviders(w http.ResponseWriter, r *http.Request, server CloudProvidersServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "GET":
			adaptCloudProvidersListRequest(w, r, server)
		default:
			errors.SendMethodNotAllowed(w, r)
			return
		}
	} else {
		switch segments[0] {
		default:
			target := server.CloudProvider(segments[0])
			if target == nil {
				errors.SendNotFound(w, r)
				return
			}
			dispatchCloudProvider(w, r, target, segments[1:])
		}
	}
}

// readCloudProvidersListRequest reads the given HTTP requests and translates it
// into an object of type CloudProvidersListServerRequest.
func readCloudProvidersListRequest(r *http.Request) (*CloudProvidersListServerRequest, error) {
	var err error
	result := new(CloudProvidersListServerRequest)
	query := r.URL.Query()
	result.order, err = helpers.ParseString(query, "order")
	if err != nil {
		return nil, err
	}
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

// writeCloudProvidersListResponse translates the given request object into an
// HTTP response.
func writeCloudProvidersListResponse(w http.ResponseWriter, r *CloudProvidersListServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptCloudProvidersListRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptCloudProvidersListRequest(w http.ResponseWriter, r *http.Request, server CloudProvidersServer) {
	request, err := readCloudProvidersListRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(CloudProvidersListServerResponse)
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
	err = writeCloudProvidersListResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
