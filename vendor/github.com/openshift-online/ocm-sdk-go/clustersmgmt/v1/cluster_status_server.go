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

// ClusterStatusServer represents the interface the manages the 'cluster_status' resource.
type ClusterStatusServer interface {

	// Get handles a request for the 'get' method.
	//
	//
	Get(ctx context.Context, request *ClusterStatusGetServerRequest, response *ClusterStatusGetServerResponse) error
}

// ClusterStatusGetServerRequest is the request for the 'get' method.
type ClusterStatusGetServerRequest struct {
}

// ClusterStatusGetServerResponse is the response for the 'get' method.
type ClusterStatusGetServerResponse struct {
	status  int
	err     *errors.Error
	status_ *ClusterStatus
}

// Status_ sets the value of the 'status' parameter.
//
//
func (r *ClusterStatusGetServerResponse) Status_(value *ClusterStatus) *ClusterStatusGetServerResponse {
	r.status_ = value
	return r
}

// Status sets the status code.
func (r *ClusterStatusGetServerResponse) Status(value int) *ClusterStatusGetServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'get' method.
func (r *ClusterStatusGetServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.status_.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// dispatchClusterStatus navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchClusterStatus(w http.ResponseWriter, r *http.Request, server ClusterStatusServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "GET":
			adaptClusterStatusGetRequest(w, r, server)
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

// readClusterStatusGetRequest reads the given HTTP requests and translates it
// into an object of type ClusterStatusGetServerRequest.
func readClusterStatusGetRequest(r *http.Request) (*ClusterStatusGetServerRequest, error) {
	var err error
	result := new(ClusterStatusGetServerRequest)
	return result, err
}

// writeClusterStatusGetResponse translates the given request object into an
// HTTP response.
func writeClusterStatusGetResponse(w http.ResponseWriter, r *ClusterStatusGetServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptClusterStatusGetRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptClusterStatusGetRequest(w http.ResponseWriter, r *http.Request, server ClusterStatusServer) {
	request, err := readClusterStatusGetRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(ClusterStatusGetServerResponse)
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
	err = writeClusterStatusGetResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
