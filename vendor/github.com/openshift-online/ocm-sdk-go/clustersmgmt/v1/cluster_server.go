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

// ClusterServer represents the interface the manages the 'cluster' resource.
type ClusterServer interface {

	// Delete handles a request for the 'delete' method.
	//
	// Deletes the cluster.
	Delete(ctx context.Context, request *ClusterDeleteServerRequest, response *ClusterDeleteServerResponse) error

	// Get handles a request for the 'get' method.
	//
	// Retrieves the details of the cluster.
	Get(ctx context.Context, request *ClusterGetServerRequest, response *ClusterGetServerResponse) error

	// Update handles a request for the 'update' method.
	//
	// Updates the cluster.
	Update(ctx context.Context, request *ClusterUpdateServerRequest, response *ClusterUpdateServerResponse) error

	// Addons returns the target 'add_on_installations' resource.
	//
	// Refrence to the resource that manages the collection of add-ons installed on this cluster.
	Addons() AddOnInstallationsServer

	// Credentials returns the target 'credentials' resource.
	//
	// Reference to the resource that manages the credentials of the cluster.
	Credentials() CredentialsServer

	// Groups returns the target 'groups' resource.
	//
	// Reference to the resource that manages the collection of groups.
	Groups() GroupsServer

	// IdentityProviders returns the target 'identity_providers' resource.
	//
	// Reference to the resource that manages the collection of identity providers.
	IdentityProviders() IdentityProvidersServer

	// Logs returns the target 'logs' resource.
	//
	// Reference to the resource that manages the collection of logs of the cluster.
	Logs() LogsServer

	// MetricQueries returns the target 'metric_queries' resource.
	//
	// Reference to the resource that manages metrics queries for the cluster.
	MetricQueries() MetricQueriesServer

	// Status returns the target 'cluster_status' resource.
	//
	// Reference to the resource that manages the detailed status of the cluster.
	Status() ClusterStatusServer
}

// ClusterDeleteServerRequest is the request for the 'delete' method.
type ClusterDeleteServerRequest struct {
}

// ClusterDeleteServerResponse is the response for the 'delete' method.
type ClusterDeleteServerResponse struct {
	status int
	err    *errors.Error
}

// Status sets the status code.
func (r *ClusterDeleteServerResponse) Status(value int) *ClusterDeleteServerResponse {
	r.status = value
	return r
}

// ClusterGetServerRequest is the request for the 'get' method.
type ClusterGetServerRequest struct {
}

// ClusterGetServerResponse is the response for the 'get' method.
type ClusterGetServerResponse struct {
	status int
	err    *errors.Error
	body   *Cluster
}

// Body sets the value of the 'body' parameter.
//
//
func (r *ClusterGetServerResponse) Body(value *Cluster) *ClusterGetServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *ClusterGetServerResponse) Status(value int) *ClusterGetServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'get' method.
func (r *ClusterGetServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// ClusterUpdateServerRequest is the request for the 'update' method.
type ClusterUpdateServerRequest struct {
	body *Cluster
}

// Body returns the value of the 'body' parameter.
//
//
func (r *ClusterUpdateServerRequest) Body() *Cluster {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
//
func (r *ClusterUpdateServerRequest) GetBody() (value *Cluster, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal request to the
// 'update' method.
func (r *ClusterUpdateServerRequest) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(clusterData)
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

// ClusterUpdateServerResponse is the response for the 'update' method.
type ClusterUpdateServerResponse struct {
	status int
	err    *errors.Error
	body   *Cluster
}

// Body sets the value of the 'body' parameter.
//
//
func (r *ClusterUpdateServerResponse) Body(value *Cluster) *ClusterUpdateServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *ClusterUpdateServerResponse) Status(value int) *ClusterUpdateServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'update' method.
func (r *ClusterUpdateServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// dispatchCluster navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchCluster(w http.ResponseWriter, r *http.Request, server ClusterServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "DELETE":
			adaptClusterDeleteRequest(w, r, server)
		case "GET":
			adaptClusterGetRequest(w, r, server)
		case "PATCH":
			adaptClusterUpdateRequest(w, r, server)
		default:
			errors.SendMethodNotAllowed(w, r)
			return
		}
	} else {
		switch segments[0] {
		case "addons":
			target := server.Addons()
			if target == nil {
				errors.SendNotFound(w, r)
				return
			}
			dispatchAddOnInstallations(w, r, target, segments[1:])
		case "credentials":
			target := server.Credentials()
			if target == nil {
				errors.SendNotFound(w, r)
				return
			}
			dispatchCredentials(w, r, target, segments[1:])
		case "groups":
			target := server.Groups()
			if target == nil {
				errors.SendNotFound(w, r)
				return
			}
			dispatchGroups(w, r, target, segments[1:])
		case "identity_providers":
			target := server.IdentityProviders()
			if target == nil {
				errors.SendNotFound(w, r)
				return
			}
			dispatchIdentityProviders(w, r, target, segments[1:])
		case "logs":
			target := server.Logs()
			if target == nil {
				errors.SendNotFound(w, r)
				return
			}
			dispatchLogs(w, r, target, segments[1:])
		case "metric_queries":
			target := server.MetricQueries()
			if target == nil {
				errors.SendNotFound(w, r)
				return
			}
			dispatchMetricQueries(w, r, target, segments[1:])
		case "status":
			target := server.Status()
			if target == nil {
				errors.SendNotFound(w, r)
				return
			}
			dispatchClusterStatus(w, r, target, segments[1:])
		default:
			errors.SendNotFound(w, r)
			return
		}
	}
}

// readClusterDeleteRequest reads the given HTTP requests and translates it
// into an object of type ClusterDeleteServerRequest.
func readClusterDeleteRequest(r *http.Request) (*ClusterDeleteServerRequest, error) {
	var err error
	result := new(ClusterDeleteServerRequest)
	return result, err
}

// writeClusterDeleteResponse translates the given request object into an
// HTTP response.
func writeClusterDeleteResponse(w http.ResponseWriter, r *ClusterDeleteServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	return nil
}

// adaptClusterDeleteRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptClusterDeleteRequest(w http.ResponseWriter, r *http.Request, server ClusterServer) {
	request, err := readClusterDeleteRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(ClusterDeleteServerResponse)
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
	err = writeClusterDeleteResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// readClusterGetRequest reads the given HTTP requests and translates it
// into an object of type ClusterGetServerRequest.
func readClusterGetRequest(r *http.Request) (*ClusterGetServerRequest, error) {
	var err error
	result := new(ClusterGetServerRequest)
	return result, err
}

// writeClusterGetResponse translates the given request object into an
// HTTP response.
func writeClusterGetResponse(w http.ResponseWriter, r *ClusterGetServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptClusterGetRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptClusterGetRequest(w http.ResponseWriter, r *http.Request, server ClusterServer) {
	request, err := readClusterGetRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(ClusterGetServerResponse)
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
	err = writeClusterGetResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// readClusterUpdateRequest reads the given HTTP requests and translates it
// into an object of type ClusterUpdateServerRequest.
func readClusterUpdateRequest(r *http.Request) (*ClusterUpdateServerRequest, error) {
	var err error
	result := new(ClusterUpdateServerRequest)
	err = result.unmarshal(r.Body)
	if err != nil {
		return nil, err
	}
	return result, err
}

// writeClusterUpdateResponse translates the given request object into an
// HTTP response.
func writeClusterUpdateResponse(w http.ResponseWriter, r *ClusterUpdateServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptClusterUpdateRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptClusterUpdateRequest(w http.ResponseWriter, r *http.Request, server ClusterServer) {
	request, err := readClusterUpdateRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(ClusterUpdateServerResponse)
	response.status = 204
	err = server.Update(r.Context(), request, response)
	if err != nil {
		glog.Errorf(
			"Can't process request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	err = writeClusterUpdateResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
