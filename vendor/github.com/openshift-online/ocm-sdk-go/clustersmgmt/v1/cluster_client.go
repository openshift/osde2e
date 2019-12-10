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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// ClusterClient is the client of the 'cluster' resource.
//
// Manages a specific cluster.
type ClusterClient struct {
	transport http.RoundTripper
	path      string
	metric    string
}

// NewClusterClient creates a new client for the 'cluster'
// resource using the given transport to sned the requests and receive the
// responses.
func NewClusterClient(transport http.RoundTripper, path string, metric string) *ClusterClient {
	client := new(ClusterClient)
	client.transport = transport
	client.path = path
	client.metric = metric
	return client
}

// Delete creates a request for the 'delete' method.
//
// Deletes the cluster.
func (c *ClusterClient) Delete() *ClusterDeleteRequest {
	request := new(ClusterDeleteRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// Get creates a request for the 'get' method.
//
// Retrieves the details of the cluster.
func (c *ClusterClient) Get() *ClusterGetRequest {
	request := new(ClusterGetRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// Update creates a request for the 'update' method.
//
// Updates the cluster.
func (c *ClusterClient) Update() *ClusterUpdateRequest {
	request := new(ClusterUpdateRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// Credentials returns the target 'credentials' resource.
//
// Reference to the resource that manages the credentials of the cluster.
func (c *ClusterClient) Credentials() *CredentialsClient {
	return NewCredentialsClient(
		c.transport,
		path.Join(c.path, "credentials"),
		path.Join(c.metric, "credentials"),
	)
}

// Groups returns the target 'groups' resource.
//
// Reference to the resource that manages the collection of groups.
func (c *ClusterClient) Groups() *GroupsClient {
	return NewGroupsClient(
		c.transport,
		path.Join(c.path, "groups"),
		path.Join(c.metric, "groups"),
	)
}

// IdentityProviders returns the target 'identity_providers' resource.
//
// Reference to the resource that manages the collection of identity providers.
func (c *ClusterClient) IdentityProviders() *IdentityProvidersClient {
	return NewIdentityProvidersClient(
		c.transport,
		path.Join(c.path, "identity_providers"),
		path.Join(c.metric, "identity_providers"),
	)
}

// Logs returns the target 'logs' resource.
//
// Reference to the resource that manages the collection of logs of the cluster.
func (c *ClusterClient) Logs() *LogsClient {
	return NewLogsClient(
		c.transport,
		path.Join(c.path, "logs"),
		path.Join(c.metric, "logs"),
	)
}

// MetricQueries returns the target 'metric_queries' resource.
//
// Reference to the resource that manages metrics queries for the cluster.
func (c *ClusterClient) MetricQueries() *MetricQueriesClient {
	return NewMetricQueriesClient(
		c.transport,
		path.Join(c.path, "metric_queries"),
		path.Join(c.metric, "metric_queries"),
	)
}

// Status returns the target 'cluster_status' resource.
//
// Reference to the resource that manages the detailed status of the cluster.
func (c *ClusterClient) Status() *ClusterStatusClient {
	return NewClusterStatusClient(
		c.transport,
		path.Join(c.path, "status"),
		path.Join(c.metric, "status"),
	)
}

// ClusterDeleteRequest is the request for the 'delete' method.
type ClusterDeleteRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
}

// Parameter adds a query parameter.
func (r *ClusterDeleteRequest) Parameter(name string, value interface{}) *ClusterDeleteRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *ClusterDeleteRequest) Header(name string, value interface{}) *ClusterDeleteRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *ClusterDeleteRequest) Send() (result *ClusterDeleteResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *ClusterDeleteRequest) SendContext(ctx context.Context) (result *ClusterDeleteResponse, err error) {
	query := helpers.CopyQuery(r.query)
	header := helpers.SetHeader(r.header, r.metric)
	uri := &url.URL{
		Path:     r.path,
		RawQuery: query.Encode(),
	}
	request := &http.Request{
		Method: "DELETE",
		URL:    uri,
		Header: header,
	}
	if ctx != nil {
		request = request.WithContext(ctx)
	}
	response, err := r.transport.RoundTrip(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	result = new(ClusterDeleteResponse)
	result.status = response.StatusCode
	result.header = response.Header
	if result.status >= 400 {
		result.err, err = errors.UnmarshalError(response.Body)
		if err != nil {
			return
		}
		err = result.err
		return
	}
	return
}

// ClusterDeleteResponse is the response for the 'delete' method.
type ClusterDeleteResponse struct {
	status int
	header http.Header
	err    *errors.Error
}

// Status returns the response status code.
func (r *ClusterDeleteResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *ClusterDeleteResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *ClusterDeleteResponse) Error() *errors.Error {
	return r.err
}

// ClusterGetRequest is the request for the 'get' method.
type ClusterGetRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
}

// Parameter adds a query parameter.
func (r *ClusterGetRequest) Parameter(name string, value interface{}) *ClusterGetRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *ClusterGetRequest) Header(name string, value interface{}) *ClusterGetRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *ClusterGetRequest) Send() (result *ClusterGetResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *ClusterGetRequest) SendContext(ctx context.Context) (result *ClusterGetResponse, err error) {
	query := helpers.CopyQuery(r.query)
	header := helpers.SetHeader(r.header, r.metric)
	uri := &url.URL{
		Path:     r.path,
		RawQuery: query.Encode(),
	}
	request := &http.Request{
		Method: "GET",
		URL:    uri,
		Header: header,
	}
	if ctx != nil {
		request = request.WithContext(ctx)
	}
	response, err := r.transport.RoundTrip(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	result = new(ClusterGetResponse)
	result.status = response.StatusCode
	result.header = response.Header
	if result.status >= 400 {
		result.err, err = errors.UnmarshalError(response.Body)
		if err != nil {
			return
		}
		err = result.err
		return
	}
	err = result.unmarshal(response.Body)
	if err != nil {
		return
	}
	return
}

// ClusterGetResponse is the response for the 'get' method.
type ClusterGetResponse struct {
	status int
	header http.Header
	err    *errors.Error
	body   *Cluster
}

// Status returns the response status code.
func (r *ClusterGetResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *ClusterGetResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *ClusterGetResponse) Error() *errors.Error {
	return r.err
}

// Body returns the value of the 'body' parameter.
//
//
func (r *ClusterGetResponse) Body() *Cluster {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
//
func (r *ClusterGetResponse) GetBody() (value *Cluster, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal responses to the
// 'get' method.
func (r *ClusterGetResponse) unmarshal(reader io.Reader) error {
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

// ClusterUpdateRequest is the request for the 'update' method.
type ClusterUpdateRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
	body      *Cluster
}

// Parameter adds a query parameter.
func (r *ClusterUpdateRequest) Parameter(name string, value interface{}) *ClusterUpdateRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *ClusterUpdateRequest) Header(name string, value interface{}) *ClusterUpdateRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Body sets the value of the 'body' parameter.
//
//
func (r *ClusterUpdateRequest) Body(value *Cluster) *ClusterUpdateRequest {
	r.body = value
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *ClusterUpdateRequest) Send() (result *ClusterUpdateResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *ClusterUpdateRequest) SendContext(ctx context.Context) (result *ClusterUpdateResponse, err error) {
	query := helpers.CopyQuery(r.query)
	header := helpers.SetHeader(r.header, r.metric)
	buffer := new(bytes.Buffer)
	err = r.marshal(buffer)
	if err != nil {
		return
	}
	uri := &url.URL{
		Path:     r.path,
		RawQuery: query.Encode(),
	}
	request := &http.Request{
		Method: "PATCH",
		URL:    uri,
		Header: header,
		Body:   ioutil.NopCloser(buffer),
	}
	if ctx != nil {
		request = request.WithContext(ctx)
	}
	response, err := r.transport.RoundTrip(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	result = new(ClusterUpdateResponse)
	result.status = response.StatusCode
	result.header = response.Header
	if result.status >= 400 {
		result.err, err = errors.UnmarshalError(response.Body)
		if err != nil {
			return
		}
		err = result.err
		return
	}
	return
}

// marshall is the method used internally to marshal requests for the
// 'update' method.
func (r *ClusterUpdateRequest) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// ClusterUpdateResponse is the response for the 'update' method.
type ClusterUpdateResponse struct {
	status int
	header http.Header
	err    *errors.Error
}

// Status returns the response status code.
func (r *ClusterUpdateResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *ClusterUpdateResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *ClusterUpdateResponse) Error() *errors.Error {
	return r.err
}
