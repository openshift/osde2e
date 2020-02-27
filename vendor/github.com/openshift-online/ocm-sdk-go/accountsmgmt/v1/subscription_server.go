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
	"net/http"

	"github.com/golang/glog"
	"github.com/openshift-online/ocm-sdk-go/errors"
)

// SubscriptionServer represents the interface the manages the 'subscription' resource.
type SubscriptionServer interface {

	// Delete handles a request for the 'delete' method.
	//
	// Deletes the subscription.
	Delete(ctx context.Context, request *SubscriptionDeleteServerRequest, response *SubscriptionDeleteServerResponse) error

	// Get handles a request for the 'get' method.
	//
	// Retrieves the details of the subscription.
	Get(ctx context.Context, request *SubscriptionGetServerRequest, response *SubscriptionGetServerResponse) error

	// ReservedResources returns the target 'subscription_reserved_resources' resource.
	//
	// Reference to the resource that manages the collection of resources reserved by the
	// subscription.
	ReservedResources() SubscriptionReservedResourcesServer
}

// SubscriptionDeleteServerRequest is the request for the 'delete' method.
type SubscriptionDeleteServerRequest struct {
}

// SubscriptionDeleteServerResponse is the response for the 'delete' method.
type SubscriptionDeleteServerResponse struct {
	status int
	err    *errors.Error
}

// Status sets the status code.
func (r *SubscriptionDeleteServerResponse) Status(value int) *SubscriptionDeleteServerResponse {
	r.status = value
	return r
}

// SubscriptionGetServerRequest is the request for the 'get' method.
type SubscriptionGetServerRequest struct {
}

// SubscriptionGetServerResponse is the response for the 'get' method.
type SubscriptionGetServerResponse struct {
	status int
	err    *errors.Error
	body   *Subscription
}

// Body sets the value of the 'body' parameter.
//
//
func (r *SubscriptionGetServerResponse) Body(value *Subscription) *SubscriptionGetServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *SubscriptionGetServerResponse) Status(value int) *SubscriptionGetServerResponse {
	r.status = value
	return r
}

// dispatchSubscription navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchSubscription(w http.ResponseWriter, r *http.Request, server SubscriptionServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "DELETE":
			adaptSubscriptionDeleteRequest(w, r, server)
			return
		case "GET":
			adaptSubscriptionGetRequest(w, r, server)
			return
		default:
			errors.SendMethodNotAllowed(w, r)
			return
		}
	}
	switch segments[0] {
	case "reserved_resources":
		target := server.ReservedResources()
		if target == nil {
			errors.SendNotFound(w, r)
			return
		}
		dispatchSubscriptionReservedResources(w, r, target, segments[1:])
	default:
		errors.SendNotFound(w, r)
		return
	}
}

// adaptSubscriptionDeleteRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptSubscriptionDeleteRequest(w http.ResponseWriter, r *http.Request, server SubscriptionServer) {
	request := &SubscriptionDeleteServerRequest{}
	err := readSubscriptionDeleteRequest(request, r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := &SubscriptionDeleteServerResponse{}
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
	err = writeSubscriptionDeleteResponse(response, w)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// adaptSubscriptionGetRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptSubscriptionGetRequest(w http.ResponseWriter, r *http.Request, server SubscriptionServer) {
	request := &SubscriptionGetServerRequest{}
	err := readSubscriptionGetRequest(request, r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := &SubscriptionGetServerResponse{}
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
	err = writeSubscriptionGetResponse(response, w)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
