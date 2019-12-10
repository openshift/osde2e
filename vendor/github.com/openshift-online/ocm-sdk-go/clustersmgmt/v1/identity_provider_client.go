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
	"net/url"

	"github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// IdentityProviderClient is the client of the 'identity_provider' resource.
//
// Manages a specific identity provider.
type IdentityProviderClient struct {
	transport http.RoundTripper
	path      string
	metric    string
}

// NewIdentityProviderClient creates a new client for the 'identity_provider'
// resource using the given transport to sned the requests and receive the
// responses.
func NewIdentityProviderClient(transport http.RoundTripper, path string, metric string) *IdentityProviderClient {
	client := new(IdentityProviderClient)
	client.transport = transport
	client.path = path
	client.metric = metric
	return client
}

// Delete creates a request for the 'delete' method.
//
// Deletes the identity provider.
func (c *IdentityProviderClient) Delete() *IdentityProviderDeleteRequest {
	request := new(IdentityProviderDeleteRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// Get creates a request for the 'get' method.
//
// Retrieves the details of the identity provider.
func (c *IdentityProviderClient) Get() *IdentityProviderGetRequest {
	request := new(IdentityProviderGetRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// IdentityProviderDeleteRequest is the request for the 'delete' method.
type IdentityProviderDeleteRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
}

// Parameter adds a query parameter.
func (r *IdentityProviderDeleteRequest) Parameter(name string, value interface{}) *IdentityProviderDeleteRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *IdentityProviderDeleteRequest) Header(name string, value interface{}) *IdentityProviderDeleteRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *IdentityProviderDeleteRequest) Send() (result *IdentityProviderDeleteResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *IdentityProviderDeleteRequest) SendContext(ctx context.Context) (result *IdentityProviderDeleteResponse, err error) {
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
	result = new(IdentityProviderDeleteResponse)
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

// IdentityProviderDeleteResponse is the response for the 'delete' method.
type IdentityProviderDeleteResponse struct {
	status int
	header http.Header
	err    *errors.Error
}

// Status returns the response status code.
func (r *IdentityProviderDeleteResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *IdentityProviderDeleteResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *IdentityProviderDeleteResponse) Error() *errors.Error {
	return r.err
}

// IdentityProviderGetRequest is the request for the 'get' method.
type IdentityProviderGetRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
}

// Parameter adds a query parameter.
func (r *IdentityProviderGetRequest) Parameter(name string, value interface{}) *IdentityProviderGetRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *IdentityProviderGetRequest) Header(name string, value interface{}) *IdentityProviderGetRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *IdentityProviderGetRequest) Send() (result *IdentityProviderGetResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *IdentityProviderGetRequest) SendContext(ctx context.Context) (result *IdentityProviderGetResponse, err error) {
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
	result = new(IdentityProviderGetResponse)
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

// IdentityProviderGetResponse is the response for the 'get' method.
type IdentityProviderGetResponse struct {
	status int
	header http.Header
	err    *errors.Error
	body   *IdentityProvider
}

// Status returns the response status code.
func (r *IdentityProviderGetResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *IdentityProviderGetResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *IdentityProviderGetResponse) Error() *errors.Error {
	return r.err
}

// Body returns the value of the 'body' parameter.
//
//
func (r *IdentityProviderGetResponse) Body() *IdentityProvider {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
//
func (r *IdentityProviderGetResponse) GetBody() (value *IdentityProvider, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal responses to the
// 'get' method.
func (r *IdentityProviderGetResponse) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(identityProviderData)
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
