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

// SubscriptionReservedResourcesServer represents the interface the manages the 'subscription_reserved_resources' resource.
type SubscriptionReservedResourcesServer interface {

	// List handles a request for the 'list' method.
	//
	// Retrieves items of the collection of reserved resources by the subscription.
	List(ctx context.Context, request *SubscriptionReservedResourcesListServerRequest, response *SubscriptionReservedResourcesListServerResponse) error

	// ReservedResource returns the target 'subscription_reserved_resource' server for the given identifier.
	//
	// Reference to the resource that manages the a specific resource reserved by a
	// subscription.
	ReservedResource(id string) SubscriptionReservedResourceServer
}

// SubscriptionReservedResourcesListServerRequest is the request for the 'list' method.
type SubscriptionReservedResourcesListServerRequest struct {
	page *int
	size *int
}

// Page returns the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
func (r *SubscriptionReservedResourcesListServerRequest) Page() int {
	if r != nil && r.page != nil {
		return *r.page
	}
	return 0
}

// GetPage returns the value of the 'page' parameter and
// a flag indicating if the parameter has a value.
//
// Index of the requested page, where one corresponds to the first page.
func (r *SubscriptionReservedResourcesListServerRequest) GetPage() (value int, ok bool) {
	ok = r != nil && r.page != nil
	if ok {
		value = *r.page
	}
	return
}

// Size returns the value of the 'size' parameter.
//
// Maximum number of items that will be contained in the returned page.
func (r *SubscriptionReservedResourcesListServerRequest) Size() int {
	if r != nil && r.size != nil {
		return *r.size
	}
	return 0
}

// GetSize returns the value of the 'size' parameter and
// a flag indicating if the parameter has a value.
//
// Maximum number of items that will be contained in the returned page.
func (r *SubscriptionReservedResourcesListServerRequest) GetSize() (value int, ok bool) {
	ok = r != nil && r.size != nil
	if ok {
		value = *r.size
	}
	return
}

// SubscriptionReservedResourcesListServerResponse is the response for the 'list' method.
type SubscriptionReservedResourcesListServerResponse struct {
	status int
	err    *errors.Error
	items  *ReservedResourceList
	page   *int
	size   *int
	total  *int
}

// Items sets the value of the 'items' parameter.
//
// Retrieved list of reserved resources.
func (r *SubscriptionReservedResourcesListServerResponse) Items(value *ReservedResourceList) *SubscriptionReservedResourcesListServerResponse {
	r.items = value
	return r
}

// Page sets the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
func (r *SubscriptionReservedResourcesListServerResponse) Page(value int) *SubscriptionReservedResourcesListServerResponse {
	r.page = &value
	return r
}

// Size sets the value of the 'size' parameter.
//
// Maximum number of items that will be contained in the returned page.
func (r *SubscriptionReservedResourcesListServerResponse) Size(value int) *SubscriptionReservedResourcesListServerResponse {
	r.size = &value
	return r
}

// Total sets the value of the 'total' parameter.
//
// Total number of items of the collection that match the search criteria,
// regardless of the size of the page.
func (r *SubscriptionReservedResourcesListServerResponse) Total(value int) *SubscriptionReservedResourcesListServerResponse {
	r.total = &value
	return r
}

// Status sets the status code.
func (r *SubscriptionReservedResourcesListServerResponse) Status(value int) *SubscriptionReservedResourcesListServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'list' method.
func (r *SubscriptionReservedResourcesListServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data := new(subscriptionReservedResourcesListServerResponseData)
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

// subscriptionReservedResourcesListServerResponseData is the structure used internally to write the request of the
// 'list' method.
type subscriptionReservedResourcesListServerResponseData struct {
	Items reservedResourceListData "json:\"items,omitempty\""
	Page  *int                     "json:\"page,omitempty\""
	Size  *int                     "json:\"size,omitempty\""
	Total *int                     "json:\"total,omitempty\""
}

// dispatchSubscriptionReservedResources navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchSubscriptionReservedResources(w http.ResponseWriter, r *http.Request, server SubscriptionReservedResourcesServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "GET":
			adaptSubscriptionReservedResourcesListRequest(w, r, server)
		default:
			errors.SendMethodNotAllowed(w, r)
			return
		}
	} else {
		switch segments[0] {
		default:
			target := server.ReservedResource(segments[0])
			if target == nil {
				errors.SendNotFound(w, r)
				return
			}
			dispatchSubscriptionReservedResource(w, r, target, segments[1:])
		}
	}
}

// readSubscriptionReservedResourcesListRequest reads the given HTTP requests and translates it
// into an object of type SubscriptionReservedResourcesListServerRequest.
func readSubscriptionReservedResourcesListRequest(r *http.Request) (*SubscriptionReservedResourcesListServerRequest, error) {
	var err error
	result := new(SubscriptionReservedResourcesListServerRequest)
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

// writeSubscriptionReservedResourcesListResponse translates the given request object into an
// HTTP response.
func writeSubscriptionReservedResourcesListResponse(w http.ResponseWriter, r *SubscriptionReservedResourcesListServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptSubscriptionReservedResourcesListRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptSubscriptionReservedResourcesListRequest(w http.ResponseWriter, r *http.Request, server SubscriptionReservedResourcesServer) {
	request, err := readSubscriptionReservedResourcesListRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(SubscriptionReservedResourcesListServerResponse)
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
	err = writeSubscriptionReservedResourcesListResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
