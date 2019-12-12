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
	"net/url"

	"github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// RoleBindingClient is the client of the 'role_binding' resource.
//
// Manages a specific role binding.
type RoleBindingClient struct {
	transport http.RoundTripper
	path      string
	metric    string
}

// NewRoleBindingClient creates a new client for the 'role_binding'
// resource using the given transport to sned the requests and receive the
// responses.
func NewRoleBindingClient(transport http.RoundTripper, path string, metric string) *RoleBindingClient {
	client := new(RoleBindingClient)
	client.transport = transport
	client.path = path
	client.metric = metric
	return client
}

// Delete creates a request for the 'delete' method.
//
// Deletes the role binding.
func (c *RoleBindingClient) Delete() *RoleBindingDeleteRequest {
	request := new(RoleBindingDeleteRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// Get creates a request for the 'get' method.
//
// Retrieves the details of the role binding.
func (c *RoleBindingClient) Get() *RoleBindingGetRequest {
	request := new(RoleBindingGetRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// RoleBindingDeleteRequest is the request for the 'delete' method.
type RoleBindingDeleteRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
}

// Parameter adds a query parameter.
func (r *RoleBindingDeleteRequest) Parameter(name string, value interface{}) *RoleBindingDeleteRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *RoleBindingDeleteRequest) Header(name string, value interface{}) *RoleBindingDeleteRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *RoleBindingDeleteRequest) Send() (result *RoleBindingDeleteResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *RoleBindingDeleteRequest) SendContext(ctx context.Context) (result *RoleBindingDeleteResponse, err error) {
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
	result = new(RoleBindingDeleteResponse)
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

// RoleBindingDeleteResponse is the response for the 'delete' method.
type RoleBindingDeleteResponse struct {
	status int
	header http.Header
	err    *errors.Error
}

// Status returns the response status code.
func (r *RoleBindingDeleteResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *RoleBindingDeleteResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *RoleBindingDeleteResponse) Error() *errors.Error {
	return r.err
}

// RoleBindingGetRequest is the request for the 'get' method.
type RoleBindingGetRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
}

// Parameter adds a query parameter.
func (r *RoleBindingGetRequest) Parameter(name string, value interface{}) *RoleBindingGetRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *RoleBindingGetRequest) Header(name string, value interface{}) *RoleBindingGetRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *RoleBindingGetRequest) Send() (result *RoleBindingGetResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *RoleBindingGetRequest) SendContext(ctx context.Context) (result *RoleBindingGetResponse, err error) {
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
	result = new(RoleBindingGetResponse)
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

// RoleBindingGetResponse is the response for the 'get' method.
type RoleBindingGetResponse struct {
	status int
	header http.Header
	err    *errors.Error
	body   *RoleBinding
}

// Status returns the response status code.
func (r *RoleBindingGetResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *RoleBindingGetResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *RoleBindingGetResponse) Error() *errors.Error {
	return r.err
}

// Body returns the value of the 'body' parameter.
//
//
func (r *RoleBindingGetResponse) Body() *RoleBinding {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
//
func (r *RoleBindingGetResponse) GetBody() (value *RoleBinding, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal responses to the
// 'get' method.
func (r *RoleBindingGetResponse) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(roleBindingData)
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
