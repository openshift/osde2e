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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// ResourceQuotaClient is the client of the 'resource_quota' resource.
//
// Manages a specific resource quota.
type ResourceQuotaClient struct {
	transport http.RoundTripper
	path      string
	metric    string
}

// NewResourceQuotaClient creates a new client for the 'resource_quota'
// resource using the given transport to sned the requests and receive the
// responses.
func NewResourceQuotaClient(transport http.RoundTripper, path string, metric string) *ResourceQuotaClient {
	client := new(ResourceQuotaClient)
	client.transport = transport
	client.path = path
	client.metric = metric
	return client
}

// Get creates a request for the 'get' method.
//
// Retrieves the details of the resource quota.
func (c *ResourceQuotaClient) Get() *ResourceQuotaGetRequest {
	request := new(ResourceQuotaGetRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// Update creates a request for the 'update' method.
//
// Updates the resource quota.
func (c *ResourceQuotaClient) Update() *ResourceQuotaUpdateRequest {
	request := new(ResourceQuotaUpdateRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// ResourceQuotaGetRequest is the request for the 'get' method.
type ResourceQuotaGetRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
}

// Parameter adds a query parameter.
func (r *ResourceQuotaGetRequest) Parameter(name string, value interface{}) *ResourceQuotaGetRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *ResourceQuotaGetRequest) Header(name string, value interface{}) *ResourceQuotaGetRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *ResourceQuotaGetRequest) Send() (result *ResourceQuotaGetResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *ResourceQuotaGetRequest) SendContext(ctx context.Context) (result *ResourceQuotaGetResponse, err error) {
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
	result = new(ResourceQuotaGetResponse)
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

// ResourceQuotaGetResponse is the response for the 'get' method.
type ResourceQuotaGetResponse struct {
	status int
	header http.Header
	err    *errors.Error
	body   *ResourceQuota
}

// Status returns the response status code.
func (r *ResourceQuotaGetResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *ResourceQuotaGetResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *ResourceQuotaGetResponse) Error() *errors.Error {
	return r.err
}

// Body returns the value of the 'body' parameter.
//
//
func (r *ResourceQuotaGetResponse) Body() *ResourceQuota {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
//
func (r *ResourceQuotaGetResponse) GetBody() (value *ResourceQuota, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal responses to the
// 'get' method.
func (r *ResourceQuotaGetResponse) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(resourceQuotaData)
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

// ResourceQuotaUpdateRequest is the request for the 'update' method.
type ResourceQuotaUpdateRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
	body      *ResourceQuota
}

// Parameter adds a query parameter.
func (r *ResourceQuotaUpdateRequest) Parameter(name string, value interface{}) *ResourceQuotaUpdateRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *ResourceQuotaUpdateRequest) Header(name string, value interface{}) *ResourceQuotaUpdateRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Body sets the value of the 'body' parameter.
//
//
func (r *ResourceQuotaUpdateRequest) Body(value *ResourceQuota) *ResourceQuotaUpdateRequest {
	r.body = value
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *ResourceQuotaUpdateRequest) Send() (result *ResourceQuotaUpdateResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *ResourceQuotaUpdateRequest) SendContext(ctx context.Context) (result *ResourceQuotaUpdateResponse, err error) {
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
	result = new(ResourceQuotaUpdateResponse)
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
func (r *ResourceQuotaUpdateRequest) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// ResourceQuotaUpdateResponse is the response for the 'update' method.
type ResourceQuotaUpdateResponse struct {
	status int
	header http.Header
	err    *errors.Error
}

// Status returns the response status code.
func (r *ResourceQuotaUpdateResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *ResourceQuotaUpdateResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *ResourceQuotaUpdateResponse) Error() *errors.Error {
	return r.err
}
