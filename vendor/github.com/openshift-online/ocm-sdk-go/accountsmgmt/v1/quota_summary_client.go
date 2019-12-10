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

// QuotaSummaryClient is the client of the 'quota_summary' resource.
//
// Manages the quota summary for an organization.
type QuotaSummaryClient struct {
	transport http.RoundTripper
	path      string
	metric    string
}

// NewQuotaSummaryClient creates a new client for the 'quota_summary'
// resource using the given transport to sned the requests and receive the
// responses.
func NewQuotaSummaryClient(transport http.RoundTripper, path string, metric string) *QuotaSummaryClient {
	client := new(QuotaSummaryClient)
	client.transport = transport
	client.path = path
	client.metric = metric
	return client
}

// List creates a request for the 'list' method.
//
// Retrieves the Quota summary.
func (c *QuotaSummaryClient) List() *QuotaSummaryListRequest {
	request := new(QuotaSummaryListRequest)
	request.transport = c.transport
	request.path = c.path
	request.metric = c.metric
	return request
}

// QuotaSummaryListRequest is the request for the 'list' method.
type QuotaSummaryListRequest struct {
	transport http.RoundTripper
	path      string
	metric    string
	query     url.Values
	header    http.Header
	page      *int
	search    *string
	size      *int
	total     *int
}

// Parameter adds a query parameter.
func (r *QuotaSummaryListRequest) Parameter(name string, value interface{}) *QuotaSummaryListRequest {
	helpers.AddValue(&r.query, name, value)
	return r
}

// Header adds a request header.
func (r *QuotaSummaryListRequest) Header(name string, value interface{}) *QuotaSummaryListRequest {
	helpers.AddHeader(&r.header, name, value)
	return r
}

// Page sets the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
//
// Default value is `1`.
func (r *QuotaSummaryListRequest) Page(value int) *QuotaSummaryListRequest {
	r.page = &value
	return r
}

// Search sets the value of the 'search' parameter.
//
// Search criteria.
//
// The syntax of this parameter is similar to the syntax of the _where_ clause
// of an SQL statement, but using the names of the attributes of the quota
// summary instead of the names of the columns of a table. For example, in order
// to retrieve the quota summary for clusters that run in multiple availability
// zones:
//
// [source,sql]
// ----
// availability_zone_type = 'multi'
// ----
//
// If the parameter isn't provided, or if the value is empty, then all the
// items that the user has permission to see will be returned.
func (r *QuotaSummaryListRequest) Search(value string) *QuotaSummaryListRequest {
	r.search = &value
	return r
}

// Size sets the value of the 'size' parameter.
//
// Maximum number of items that will be contained in the returned page.
//
// Default value is `100`.
func (r *QuotaSummaryListRequest) Size(value int) *QuotaSummaryListRequest {
	r.size = &value
	return r
}

// Total sets the value of the 'total' parameter.
//
// Total number of items of the collection that match the search criteria,
// regardless of the size of the page.
func (r *QuotaSummaryListRequest) Total(value int) *QuotaSummaryListRequest {
	r.total = &value
	return r
}

// Send sends this request, waits for the response, and returns it.
//
// This is a potentially lengthy operation, as it requires network communication.
// Consider using a context and the SendContext method.
func (r *QuotaSummaryListRequest) Send() (result *QuotaSummaryListResponse, err error) {
	return r.SendContext(context.Background())
}

// SendContext sends this request, waits for the response, and returns it.
func (r *QuotaSummaryListRequest) SendContext(ctx context.Context) (result *QuotaSummaryListResponse, err error) {
	query := helpers.CopyQuery(r.query)
	if r.page != nil {
		helpers.AddValue(&query, "page", *r.page)
	}
	if r.search != nil {
		helpers.AddValue(&query, "search", *r.search)
	}
	if r.size != nil {
		helpers.AddValue(&query, "size", *r.size)
	}
	if r.total != nil {
		helpers.AddValue(&query, "total", *r.total)
	}
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
	result = new(QuotaSummaryListResponse)
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

// QuotaSummaryListResponse is the response for the 'list' method.
type QuotaSummaryListResponse struct {
	status int
	header http.Header
	err    *errors.Error
	items  *QuotaSummaryList
	page   *int
	size   *int
	total  *int
}

// Status returns the response status code.
func (r *QuotaSummaryListResponse) Status() int {
	return r.status
}

// Header returns header of the response.
func (r *QuotaSummaryListResponse) Header() http.Header {
	return r.header
}

// Error returns the response error.
func (r *QuotaSummaryListResponse) Error() *errors.Error {
	return r.err
}

// Items returns the value of the 'items' parameter.
//
// Retrieved quota summary items.
func (r *QuotaSummaryListResponse) Items() *QuotaSummaryList {
	if r == nil {
		return nil
	}
	return r.items
}

// GetItems returns the value of the 'items' parameter and
// a flag indicating if the parameter has a value.
//
// Retrieved quota summary items.
func (r *QuotaSummaryListResponse) GetItems() (value *QuotaSummaryList, ok bool) {
	ok = r != nil && r.items != nil
	if ok {
		value = r.items
	}
	return
}

// Page returns the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
//
// Default value is `1`.
func (r *QuotaSummaryListResponse) Page() int {
	if r != nil && r.page != nil {
		return *r.page
	}
	return 0
}

// GetPage returns the value of the 'page' parameter and
// a flag indicating if the parameter has a value.
//
// Index of the requested page, where one corresponds to the first page.
//
// Default value is `1`.
func (r *QuotaSummaryListResponse) GetPage() (value int, ok bool) {
	ok = r != nil && r.page != nil
	if ok {
		value = *r.page
	}
	return
}

// Size returns the value of the 'size' parameter.
//
// Maximum number of items that will be contained in the returned page.
//
// Default value is `100`.
func (r *QuotaSummaryListResponse) Size() int {
	if r != nil && r.size != nil {
		return *r.size
	}
	return 0
}

// GetSize returns the value of the 'size' parameter and
// a flag indicating if the parameter has a value.
//
// Maximum number of items that will be contained in the returned page.
//
// Default value is `100`.
func (r *QuotaSummaryListResponse) GetSize() (value int, ok bool) {
	ok = r != nil && r.size != nil
	if ok {
		value = *r.size
	}
	return
}

// Total returns the value of the 'total' parameter.
//
// Total number of items of the collection that match the search criteria,
// regardless of the size of the page.
func (r *QuotaSummaryListResponse) Total() int {
	if r != nil && r.total != nil {
		return *r.total
	}
	return 0
}

// GetTotal returns the value of the 'total' parameter and
// a flag indicating if the parameter has a value.
//
// Total number of items of the collection that match the search criteria,
// regardless of the size of the page.
func (r *QuotaSummaryListResponse) GetTotal() (value int, ok bool) {
	ok = r != nil && r.total != nil
	if ok {
		value = *r.total
	}
	return
}

// unmarshal is the method used internally to unmarshal responses to the
// 'list' method.
func (r *QuotaSummaryListResponse) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(quotaSummaryListResponseData)
	err = decoder.Decode(data)
	if err != nil {
		return err
	}
	r.items, err = data.Items.unwrap()
	if err != nil {
		return err
	}
	r.page = data.Page
	r.size = data.Size
	r.total = data.Total
	return err
}

// quotaSummaryListResponseData is the structure used internally to unmarshal
// the response of the 'list' method.
type quotaSummaryListResponseData struct {
	Items quotaSummaryListData "json:\"items,omitempty\""
	Page  *int                 "json:\"page,omitempty\""
	Size  *int                 "json:\"size,omitempty\""
	Total *int                 "json:\"total,omitempty\""
}
