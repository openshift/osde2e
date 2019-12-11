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
	"path"

	"github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// OrganizationClient is the client of the 'organization' resource.
//
// Manages a specific organization.
type OrganizationClient struct {
	transport http.RoundTripper
	path      string
	metric    string
}

// NewOrganizationClient creates a new client for the 'organization'
// resource using the given transport to sned the requests and receive the
// responses.
func NewOrganizationClient(transport http.RoundTripper, path string, metric string) *OrganizationClient {
	client := new(OrganizationClient)
	client.transport = transport
	client.path = path
	client.metric = metric
	return client
}

// Get creates a request for the 'get' method.
//
// Retrieves the details of the organization.
func (c *OrganizationClient) Get() *OrganizationGetRequest {
	request := new(OrganizationGetRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// Update creates a request for the 'update' method.
//
// Updates the organization.
func (c *OrganizationClient) Update() *OrganizationUpdateRequest {
	request := new(OrganizationUpdateRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// QuotaSummary returns the target 'quota_summary' resource.
//
// Reference to the service that returns the summary of the resource quota for this
// organization.
func (c *OrganizationClient) QuotaSummary() *QuotaSummaryClient {
	return NewQuotaSummaryClient(
		c.transport,
		path.Join(c.path, "quota_summary"),
		path.Join(c.metric, "quota_summary"),
	)
}

// ResourceQuota returns the target 'resource_quotas' resource.
//
// Reference to the service that manages the resource quotas for this
// organization.
func (c *OrganizationClient) ResourceQuota() *ResourceQuotasClient {
	return NewResourceQuotasClient(
		c.transport,
		path.Join(c.path, "resource_quota"),
		path.Join(c.metric, "resource_quota"),
	)
}

// OrganizationGetRequest is the request for the 'get' method.
type OrganizationGetRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
}

// Parameter adds a query parameter.
func (r *OrganizationGetRequest) Parameter(name string, value interface{}) *OrganizationGetRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *OrganizationGetRequest) Header(name string, value interface{}) *OrganizationGetRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *OrganizationGetRequest) Send() (result *OrganizationGetResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *OrganizationGetRequest) SendContext(ctx context.Context) (result *OrganizationGetResponse, err error) {
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
	result = new(OrganizationGetResponse)
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

// OrganizationGetResponse is the response for the 'get' method.
type OrganizationGetResponse struct {
	status int
	header http.Header
	err    *errors.Error
	body   *Organization
}

// Status returns the response status code.
func (r *OrganizationGetResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *OrganizationGetResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *OrganizationGetResponse) Error() *errors.Error {
	return r.err
}

// Body returns the value of the 'body' parameter.
//
//
func (r *OrganizationGetResponse) Body() *Organization {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
//
func (r *OrganizationGetResponse) GetBody() (value *Organization, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal responses to the
// 'get' method.
func (r *OrganizationGetResponse) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(organizationData)
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

// OrganizationUpdateRequest is the request for the 'update' method.
type OrganizationUpdateRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
	body      *Organization
}

// Parameter adds a query parameter.
func (r *OrganizationUpdateRequest) Parameter(name string, value interface{}) *OrganizationUpdateRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *OrganizationUpdateRequest) Header(name string, value interface{}) *OrganizationUpdateRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Body sets the value of the 'body' parameter.
//
//
func (r *OrganizationUpdateRequest) Body(value *Organization) *OrganizationUpdateRequest {
	r.body = value
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *OrganizationUpdateRequest) Send() (result *OrganizationUpdateResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *OrganizationUpdateRequest) SendContext(ctx context.Context) (result *OrganizationUpdateResponse, err error) {
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
	result = new(OrganizationUpdateResponse)
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
func (r *OrganizationUpdateRequest) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// OrganizationUpdateResponse is the response for the 'update' method.
type OrganizationUpdateResponse struct {
	status int
	header http.Header
	err    *errors.Error
}

// Status returns the response status code.
func (r *OrganizationUpdateResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *OrganizationUpdateResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *OrganizationUpdateResponse) Error() *errors.Error {
	return r.err
}
