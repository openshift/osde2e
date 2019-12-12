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

// SubscriptionClient is the client of the 'subscription' resource.
//
// Manages a specific subscription.
type SubscriptionClient struct {
	transport http.RoundTripper
	path      string
	metric    string
}

// NewSubscriptionClient creates a new client for the 'subscription'
// resource using the given transport to sned the requests and receive the
// responses.
func NewSubscriptionClient(transport http.RoundTripper, path string, metric string) *SubscriptionClient {
	client := new(SubscriptionClient)
	client.transport = transport
	client.path = path
	client.metric = metric
	return client
}

// Delete creates a request for the 'delete' method.
//
// Deletes the subscription.
func (c *SubscriptionClient) Delete() *SubscriptionDeleteRequest {
	request := new(SubscriptionDeleteRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// Get creates a request for the 'get' method.
//
// Retrieves the details of the subscription.
func (c *SubscriptionClient) Get() *SubscriptionGetRequest {
	request := new(SubscriptionGetRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// SubscriptionDeleteRequest is the request for the 'delete' method.
type SubscriptionDeleteRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
}

// Parameter adds a query parameter.
func (r *SubscriptionDeleteRequest) Parameter(name string, value interface{}) *SubscriptionDeleteRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *SubscriptionDeleteRequest) Header(name string, value interface{}) *SubscriptionDeleteRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *SubscriptionDeleteRequest) Send() (result *SubscriptionDeleteResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *SubscriptionDeleteRequest) SendContext(ctx context.Context) (result *SubscriptionDeleteResponse, err error) {
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
	result = new(SubscriptionDeleteResponse)
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

// SubscriptionDeleteResponse is the response for the 'delete' method.
type SubscriptionDeleteResponse struct {
	status int
	header http.Header
	err    *errors.Error
}

// Status returns the response status code.
func (r *SubscriptionDeleteResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *SubscriptionDeleteResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *SubscriptionDeleteResponse) Error() *errors.Error {
	return r.err
}

// SubscriptionGetRequest is the request for the 'get' method.
type SubscriptionGetRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
}

// Parameter adds a query parameter.
func (r *SubscriptionGetRequest) Parameter(name string, value interface{}) *SubscriptionGetRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *SubscriptionGetRequest) Header(name string, value interface{}) *SubscriptionGetRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *SubscriptionGetRequest) Send() (result *SubscriptionGetResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *SubscriptionGetRequest) SendContext(ctx context.Context) (result *SubscriptionGetResponse, err error) {
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
	result = new(SubscriptionGetResponse)
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

// SubscriptionGetResponse is the response for the 'get' method.
type SubscriptionGetResponse struct {
	status int
	header http.Header
	err    *errors.Error
	body   *Subscription
}

// Status returns the response status code.
func (r *SubscriptionGetResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *SubscriptionGetResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *SubscriptionGetResponse) Error() *errors.Error {
	return r.err
}

// Body returns the value of the 'body' parameter.
//
//
func (r *SubscriptionGetResponse) Body() *Subscription {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
//
func (r *SubscriptionGetResponse) GetBody() (value *Subscription, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal responses to the
// 'get' method.
func (r *SubscriptionGetResponse) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(subscriptionData)
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
