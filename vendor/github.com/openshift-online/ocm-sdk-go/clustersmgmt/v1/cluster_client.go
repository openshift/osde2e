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
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	jsoniter "github.com/json-iterator/go"
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
// resource using the given transport to send the requests and receive the
// responses.
func NewClusterClient(transport http.RoundTripper, path string, metric string) *ClusterClient {
	return &ClusterClient{
		transport: transport,
		path:      path,
		metric:    metric,
	}
}

// Delete creates a request for the 'delete' method.
//
// Deletes the cluster.
func (c *ClusterClient) Delete() *ClusterDeleteRequest {
	return &ClusterDeleteRequest{
		transport: c.transport,
		path:      c.path,
		metric:    c.metric,
	}
}

// Get creates a request for the 'get' method.
//
// Retrieves the details of the cluster.
func (c *ClusterClient) Get() *ClusterGetRequest {
	return &ClusterGetRequest{
		transport: c.transport,
		path:      c.path,
		metric:    c.metric,
	}
}

// Update creates a request for the 'update' method.
//
// Updates the cluster.
func (c *ClusterClient) Update() *ClusterUpdateRequest {
	return &ClusterUpdateRequest{
		transport: c.transport,
		path:      c.path,
		metric:    c.metric,
	}
}

// AWSInfrastructureAccessRoleGrants returns the target 'AWS_infrastructure_access_role_grants' resource.
//
// Refrence to the resource that manages the collection of AWS infrastructure
// access role grants on this cluster.
func (c *ClusterClient) AWSInfrastructureAccessRoleGrants() *AWSInfrastructureAccessRoleGrantsClient {
	return NewAWSInfrastructureAccessRoleGrantsClient(
		c.transport,
		path.Join(c.path, "aws_infrastructure_access_role_grants"),
		path.Join(c.metric, "aws_infrastructure_access_role_grants"),
	)
}

// Addons returns the target 'add_on_installations' resource.
//
// Refrence to the resource that manages the collection of add-ons installed on this cluster.
func (c *ClusterClient) Addons() *AddOnInstallationsClient {
	return NewAddOnInstallationsClient(
		c.transport,
		path.Join(c.path, "addons"),
		path.Join(c.metric, "addons"),
	)
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

// ClusterPollRequest is the request for the Poll method.
type ClusterPollRequest struct {
	request    *ClusterGetRequest
	interval   time.Duration
	statuses   []int
	predicates []func(interface{}) bool
}

// Parameter adds a query parameter to all the requests that will be used to retrieve the object.
func (r *ClusterPollRequest) Parameter(name string, value interface{}) *ClusterPollRequest {
	r.request.Parameter(name, value)
	return r
}

// Header adds a request header to all the requests that will be used to retrieve the object.
func (r *ClusterPollRequest) Header(name string, value interface{}) *ClusterPollRequest {
	r.request.Header(name, value)
	return r
}

// Interval sets the polling interval. This parameter is mandatory and must be greater than zero.
func (r *ClusterPollRequest) Interval(value time.Duration) *ClusterPollRequest {
	r.interval = value
	return r
}

// Status set the expected status of the response. Multiple values can be set calling this method
// multiple times. The response will be considered successful if the status is any of those values.
func (r *ClusterPollRequest) Status(value int) *ClusterPollRequest {
	r.statuses = append(r.statuses, value)
	return r
}

// Predicate adds a predicate that the response should satisfy be considered successful. Multiple
// predicates can be set calling this method multiple times. The response will be considered successful
// if all the predicates are satisfied.
func (r *ClusterPollRequest) Predicate(value func(*ClusterGetResponse) bool) *ClusterPollRequest {
	r.predicates = append(r.predicates, func(response interface{}) bool {
		return value(response.(*ClusterGetResponse))
	})
	return r
}

// StartContext starts the polling loop. Responses will be considered successful if the status is one of
// the values specified with the Status method and if all the predicates specified with the Predicate
// method return nil.
//
// The context must have a timeout or deadline, otherwise this method will immediately return an error.
func (r *ClusterPollRequest) StartContext(ctx context.Context) (response *ClusterPollResponse, err error) {
	result, err := helpers.PollContext(ctx, r.interval, r.statuses, r.predicates, r.task)
	if result != nil {
		response = &ClusterPollResponse{
			response: result.(*ClusterGetResponse),
		}
	}
	return
}

// task adapts the types of the request/response types so that they can be used with the generic
// polling function from the helpers package.
func (r *ClusterPollRequest) task(ctx context.Context) (status int, result interface{}, err error) {
	response, err := r.request.SendContext(ctx)
	if response != nil {
		status = response.Status()
		result = response
	}
	return
}

// ClusterPollResponse is the response for the Poll method.
type ClusterPollResponse struct {
	response *ClusterGetResponse
}

// Status returns the response status code.
func (r *ClusterPollResponse) Status() int {
	if r == nil {
		return 0
	}
	return r.response.Status()
}

// Header returns header of the response.
func (r *ClusterPollResponse) Header() http.Header {
	if r == nil {
		return nil
	}
	return r.response.Header()
}

// Error returns the response error.
func (r *ClusterPollResponse) Error() *errors.Error {
	if r == nil {
		return nil
	}
	return r.response.Error()
}

// Body returns the value of the 'body' parameter.
//
//
func (r *ClusterPollResponse) Body() *Cluster {
	return r.response.Body()
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
//
func (r *ClusterPollResponse) GetBody() (value *Cluster, ok bool) {
	return r.response.GetBody()
}

// Poll creates a request to repeatedly retrieve the object till the response has one of a given set
// of states and satisfies a set of predicates.
func (c *ClusterClient) Poll() *ClusterPollRequest {
	return &ClusterPollRequest{
		request: c.Get(),
	}
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
	result = &ClusterDeleteResponse{}
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
	if r == nil {
		return 0
	}
	return r.status
}

// Header returns header of the response.
func (r *ClusterDeleteResponse) Header() http.Header {
	if r == nil {
		return nil
	}
	return r.header
}

// Error returns the response error.
func (r *ClusterDeleteResponse) Error() *errors.Error {
	if r == nil {
		return nil
	}
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
	result = &ClusterGetResponse{}
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
	err = readClusterGetResponse(result, response.Body)
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
	if r == nil {
		return 0
	}
	return r.status
}

// Header returns header of the response.
func (r *ClusterGetResponse) Header() http.Header {
	if r == nil {
		return nil
	}
	return r.header
}

// Error returns the response error.
func (r *ClusterGetResponse) Error() *errors.Error {
	if r == nil {
		return nil
	}
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
	buffer := &bytes.Buffer{}
	err = writeClusterUpdateRequest(r, buffer)
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
	result = &ClusterUpdateResponse{}
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
	err = readClusterUpdateResponse(result, response.Body)
	if err != nil {
		return
	}
	return
}

// marshall is the method used internally to marshal requests for the
// 'update' method.
func (r *ClusterUpdateRequest) marshal(writer io.Writer) error {
	stream := helpers.NewStream(writer)
	r.stream(stream)
	return stream.Error
}
func (r *ClusterUpdateRequest) stream(stream *jsoniter.Stream) {
}

// ClusterUpdateResponse is the response for the 'update' method.
type ClusterUpdateResponse struct {
	status int
	header http.Header
	err    *errors.Error
	body   *Cluster
}

// Status returns the response status code.
func (r *ClusterUpdateResponse) Status() int {
	if r == nil {
		return 0
	}
	return r.status
}

// Header returns header of the response.
func (r *ClusterUpdateResponse) Header() http.Header {
	if r == nil {
		return nil
	}
	return r.header
}

// Error returns the response error.
func (r *ClusterUpdateResponse) Error() *errors.Error {
	if r == nil {
		return nil
	}
	return r.err
}

// Body returns the value of the 'body' parameter.
//
//
func (r *ClusterUpdateResponse) Body() *Cluster {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
//
func (r *ClusterUpdateResponse) GetBody() (value *Cluster, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}
