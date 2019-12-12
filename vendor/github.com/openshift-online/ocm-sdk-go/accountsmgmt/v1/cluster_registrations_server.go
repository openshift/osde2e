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

// ClusterRegistrationsServer represents the interface the manages the 'cluster_registrations' resource.
type ClusterRegistrationsServer interface {

	// Post handles a request for the 'post' method.
	//
	// Finds or creates a cluster registration with a registry credential
	// token and cluster identifier.
	Post(ctx context.Context, request *ClusterRegistrationsPostServerRequest, response *ClusterRegistrationsPostServerResponse) error
}

// ClusterRegistrationsPostServerRequest is the request for the 'post' method.
type ClusterRegistrationsPostServerRequest struct {
	request *ClusterRegistrationRequest
}

// Request returns the value of the 'request' parameter.
//
//
func (r *ClusterRegistrationsPostServerRequest) Request() *ClusterRegistrationRequest {
	if r == nil {
		return nil
	}
	return r.request
}

// GetRequest returns the value of the 'request' parameter and
// a flag indicating if the parameter has a value.
//
//
func (r *ClusterRegistrationsPostServerRequest) GetRequest() (value *ClusterRegistrationRequest, ok bool) {
	ok = r != nil && r.request != nil
	if ok {
		value = r.request
	}
	return
}

// unmarshal is the method used internally to unmarshal request to the
// 'post' method.
func (r *ClusterRegistrationsPostServerRequest) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(clusterRegistrationRequestData)
	err = decoder.Decode(data)
	if err != nil {
		return err
	}
	r.request, err = data.unwrap()
	if err != nil {
		return err
	}
	return err
}

// ClusterRegistrationsPostServerResponse is the response for the 'post' method.
type ClusterRegistrationsPostServerResponse struct {
	status   int
	err      *errors.Error
	response *ClusterRegistrationResponse
}

// Response sets the value of the 'response' parameter.
//
//
func (r *ClusterRegistrationsPostServerResponse) Response(value *ClusterRegistrationResponse) *ClusterRegistrationsPostServerResponse {
	r.response = value
	return r
}

// Status sets the status code.
func (r *ClusterRegistrationsPostServerResponse) Status(value int) *ClusterRegistrationsPostServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'post' method.
func (r *ClusterRegistrationsPostServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.response.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// dispatchClusterRegistrations navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchClusterRegistrations(w http.ResponseWriter, r *http.Request, server ClusterRegistrationsServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "POST":
			adaptClusterRegistrationsPostRequest(w, r, server)
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

// readClusterRegistrationsPostRequest reads the given HTTP requests and translates it
// into an object of type ClusterRegistrationsPostServerRequest.
func readClusterRegistrationsPostRequest(r *http.Request) (*ClusterRegistrationsPostServerRequest, error) {
	var err error
	result := new(ClusterRegistrationsPostServerRequest)
	err = result.unmarshal(r.Body)
	if err != nil {
		return nil, err
	}
	return result, err
}

// writeClusterRegistrationsPostResponse translates the given request object into an
// HTTP response.
func writeClusterRegistrationsPostResponse(w http.ResponseWriter, r *ClusterRegistrationsPostServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptClusterRegistrationsPostRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptClusterRegistrationsPostRequest(w http.ResponseWriter, r *http.Request, server ClusterRegistrationsServer) {
	request, err := readClusterRegistrationsPostRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(ClusterRegistrationsPostServerResponse)
	response.status = 201
	err = server.Post(r.Context(), request, response)
	if err != nil {
		glog.Errorf(
			"Can't process request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	err = writeClusterRegistrationsPostResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
