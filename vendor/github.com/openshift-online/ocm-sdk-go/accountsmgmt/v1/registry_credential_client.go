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

// RegistryCredentialClient is the client of the 'registry_credential' resource.
//
// Manages a specific registry credential.
type RegistryCredentialClient struct {
	transport http.RoundTripper
	path      string
	metric    string
}

// NewRegistryCredentialClient creates a new client for the 'registry_credential'
// resource using the given transport to sned the requests and receive the
// responses.
func NewRegistryCredentialClient(transport http.RoundTripper, path string, metric string) *RegistryCredentialClient {
	client := new(RegistryCredentialClient)
	client.transport = transport
	client.path = path
	client.metric = metric
	return client
}

// Get creates a request for the 'get' method.
//
// Retrieves the details of the registry credential.
func (c *RegistryCredentialClient) Get() *RegistryCredentialGetRequest {
	request := new(RegistryCredentialGetRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// RegistryCredentialGetRequest is the request for the 'get' method.
type RegistryCredentialGetRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
}

// Parameter adds a query parameter.
func (r *RegistryCredentialGetRequest) Parameter(name string, value interface{}) *RegistryCredentialGetRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *RegistryCredentialGetRequest) Header(name string, value interface{}) *RegistryCredentialGetRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *RegistryCredentialGetRequest) Send() (result *RegistryCredentialGetResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *RegistryCredentialGetRequest) SendContext(ctx context.Context) (result *RegistryCredentialGetResponse, err error) {
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
	result = new(RegistryCredentialGetResponse)
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

// RegistryCredentialGetResponse is the response for the 'get' method.
type RegistryCredentialGetResponse struct {
	status int
	header http.Header
	err    *errors.Error
	body   *RegistryCredential
}

// Status returns the response status code.
func (r *RegistryCredentialGetResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *RegistryCredentialGetResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *RegistryCredentialGetResponse) Error() *errors.Error {
	return r.err
}

// Body returns the value of the 'body' parameter.
//
//
func (r *RegistryCredentialGetResponse) Body() *RegistryCredential {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
//
func (r *RegistryCredentialGetResponse) GetBody() (value *RegistryCredential, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal responses to the
// 'get' method.
func (r *RegistryCredentialGetResponse) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(registryCredentialData)
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
