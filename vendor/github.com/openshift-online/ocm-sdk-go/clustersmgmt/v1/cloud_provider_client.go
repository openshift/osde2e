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
	"path"

	"github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// CloudProviderClient is the client of the 'cloud_provider' resource.
//
// Manages a specific cloud provider.
type CloudProviderClient struct {
	transport http.RoundTripper
	path      string
	metric    string
}

// NewCloudProviderClient creates a new client for the 'cloud_provider'
// resource using the given transport to sned the requests and receive the
// responses.
func NewCloudProviderClient(transport http.RoundTripper, path string, metric string) *CloudProviderClient {
	client := new(CloudProviderClient)
	client.transport = transport
	client.path = path
	client.metric = metric
	return client
}

// Get creates a request for the 'get' method.
//
// Retrieves the details of the cloud provider.
func (c *CloudProviderClient) Get() *CloudProviderGetRequest {
	request := new(CloudProviderGetRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// Regions returns the target 'cloud_regions' resource.
//
// Reference to the resource that manages the collection of regions for
// this cloud provider.
func (c *CloudProviderClient) Regions() *CloudRegionsClient {
	return NewCloudRegionsClient(
		c.transport,
		path.Join(c.path, "regions"),
		path.Join(c.metric, "regions"),
	)
}

// CloudProviderGetRequest is the request for the 'get' method.
type CloudProviderGetRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
}

// Parameter adds a query parameter.
func (r *CloudProviderGetRequest) Parameter(name string, value interface{}) *CloudProviderGetRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *CloudProviderGetRequest) Header(name string, value interface{}) *CloudProviderGetRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *CloudProviderGetRequest) Send() (result *CloudProviderGetResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *CloudProviderGetRequest) SendContext(ctx context.Context) (result *CloudProviderGetResponse, err error) {
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
	result = new(CloudProviderGetResponse)
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

// CloudProviderGetResponse is the response for the 'get' method.
type CloudProviderGetResponse struct {
	status int
	header http.Header
	err    *errors.Error
	body   *CloudProvider
}

// Status returns the response status code.
func (r *CloudProviderGetResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *CloudProviderGetResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *CloudProviderGetResponse) Error() *errors.Error {
	return r.err
}

// Body returns the value of the 'body' parameter.
//
//
func (r *CloudProviderGetResponse) Body() *CloudProvider {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
//
func (r *CloudProviderGetResponse) GetBody() (value *CloudProvider, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal responses to the
// 'get' method.
func (r *CloudProviderGetResponse) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(cloudProviderData)
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
