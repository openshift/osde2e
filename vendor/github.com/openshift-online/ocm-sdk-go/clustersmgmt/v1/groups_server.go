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

// GroupsServer represents the interface the manages the 'groups' resource.
type GroupsServer interface {

	// List handles a request for the 'list' method.
	//
	// Retrieves the list of groups.
	List(ctx context.Context, request *GroupsListServerRequest, response *GroupsListServerResponse) error

	// Group returns the target 'group' server for the given identifier.
	//
	// Reference to the service that manages an specific group.
	Group(id string) GroupServer
}

// GroupsListServerRequest is the request for the 'list' method.
type GroupsListServerRequest struct {
}

// GroupsListServerResponse is the response for the 'list' method.
type GroupsListServerResponse struct {
	status int
	err    *errors.Error
	items  *GroupList
	page   *int
	size   *int
	total  *int
}

// Items sets the value of the 'items' parameter.
//
// Retrieved list of groups.
func (r *GroupsListServerResponse) Items(value *GroupList) *GroupsListServerResponse {
	r.items = value
	return r
}

// Page sets the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
func (r *GroupsListServerResponse) Page(value int) *GroupsListServerResponse {
	r.page = &value
	return r
}

// Size sets the value of the 'size' parameter.
//
// Number of items contained in the returned page.
func (r *GroupsListServerResponse) Size(value int) *GroupsListServerResponse {
	r.size = &value
	return r
}

// Total sets the value of the 'total' parameter.
//
// Total number of items of the collection.
func (r *GroupsListServerResponse) Total(value int) *GroupsListServerResponse {
	r.total = &value
	return r
}

// Status sets the status code.
func (r *GroupsListServerResponse) Status(value int) *GroupsListServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'list' method.
func (r *GroupsListServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data := new(groupsListServerResponseData)
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

// groupsListServerResponseData is the structure used internally to write the request of the
// 'list' method.
type groupsListServerResponseData struct {
	Items groupListData "json:\"items,omitempty\""
	Page  *int          "json:\"page,omitempty\""
	Size  *int          "json:\"size,omitempty\""
	Total *int          "json:\"total,omitempty\""
}

// dispatchGroups navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchGroups(w http.ResponseWriter, r *http.Request, server GroupsServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "GET":
			adaptGroupsListRequest(w, r, server)
		default:
			errors.SendMethodNotAllowed(w, r)
			return
		}
	} else {
		switch segments[0] {
		default:
			target := server.Group(segments[0])
			if target == nil {
				errors.SendNotFound(w, r)
				return
			}
			dispatchGroup(w, r, target, segments[1:])
		}
	}
}

// readGroupsListRequest reads the given HTTP requests and translates it
// into an object of type GroupsListServerRequest.
func readGroupsListRequest(r *http.Request) (*GroupsListServerRequest, error) {
	var err error
	result := new(GroupsListServerRequest)
	return result, err
}

// writeGroupsListResponse translates the given request object into an
// HTTP response.
func writeGroupsListResponse(w http.ResponseWriter, r *GroupsListServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptGroupsListRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptGroupsListRequest(w http.ResponseWriter, r *http.Request, server GroupsServer) {
	request, err := readGroupsListRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(GroupsListServerResponse)
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
	err = writeGroupsListResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
