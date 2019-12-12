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

	"github.com/golang/glog"
	"github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// AccountsServer represents the interface the manages the 'accounts' resource.
type AccountsServer interface {

	// Add handles a request for the 'add' method.
	//
	// Creates a new account.
	Add(ctx context.Context, request *AccountsAddServerRequest, response *AccountsAddServerResponse) error

	// List handles a request for the 'list' method.
	//
	// Retrieves the list of accounts.
	List(ctx context.Context, request *AccountsListServerRequest, response *AccountsListServerResponse) error

	// Account returns the target 'account' server for the given identifier.
	//
	// Reference to the service that manages an specific account.
	Account(id string) AccountServer
}

// AccountsAddServerRequest is the request for the 'add' method.
type AccountsAddServerRequest struct {
	body *Account
}

// Body returns the value of the 'body' parameter.
//
// Account data.
func (r *AccountsAddServerRequest) Body() *Account {
	if r == nil {
		return nil
	}
	return r.body
}

// GetBody returns the value of the 'body' parameter and
// a flag indicating if the parameter has a value.
//
// Account data.
func (r *AccountsAddServerRequest) GetBody() (value *Account, ok bool) {
	ok = r != nil && r.body != nil
	if ok {
		value = r.body
	}
	return
}

// unmarshal is the method used internally to unmarshal request to the
// 'add' method.
func (r *AccountsAddServerRequest) unmarshal(reader io.Reader) error {
	var err error
	decoder := json.NewDecoder(reader)
	data := new(accountData)
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

// AccountsAddServerResponse is the response for the 'add' method.
type AccountsAddServerResponse struct {
	status int
	err    *errors.Error
	body   *Account
}

// Body sets the value of the 'body' parameter.
//
// Account data.
func (r *AccountsAddServerResponse) Body(value *Account) *AccountsAddServerResponse {
	r.body = value
	return r
}

// Status sets the status code.
func (r *AccountsAddServerResponse) Status(value int) *AccountsAddServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'add' method.
func (r *AccountsAddServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data, err := r.body.wrap()
	if err != nil {
		return err
	}
	err = encoder.Encode(data)
	return err
}

// AccountsListServerRequest is the request for the 'list' method.
type AccountsListServerRequest struct {
	order  *string
	page   *int
	search *string
	size   *int
}

// Order returns the value of the 'order' parameter.
//
// Order criteria.
//
// The syntax of this parameter is similar to the syntax of the _order by_ clause of
// a SQL statement. For example, in order to sort the
// accounts descending by name identifier the value should be:
//
// [source,sql]
// ----
// name desc
// ----
//
// If the parameter isn't provided, or if the value is empty, then the order of the
// results is undefined.
func (r *AccountsListServerRequest) Order() string {
	if r != nil && r.order != nil {
		return *r.order
	}
	return ""
}

// GetOrder returns the value of the 'order' parameter and
// a flag indicating if the parameter has a value.
//
// Order criteria.
//
// The syntax of this parameter is similar to the syntax of the _order by_ clause of
// a SQL statement. For example, in order to sort the
// accounts descending by name identifier the value should be:
//
// [source,sql]
// ----
// name desc
// ----
//
// If the parameter isn't provided, or if the value is empty, then the order of the
// results is undefined.
func (r *AccountsListServerRequest) GetOrder() (value string, ok bool) {
	ok = r != nil && r.order != nil
	if ok {
		value = *r.order
	}
	return
}

// Page returns the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
func (r *AccountsListServerRequest) Page() int {
	if r != nil && r.page != nil {
		return *r.page
	}
	return 0
}

// GetPage returns the value of the 'page' parameter and
// a flag indicating if the parameter has a value.
//
// Index of the requested page, where one corresponds to the first page.
func (r *AccountsListServerRequest) GetPage() (value int, ok bool) {
	ok = r != nil && r.page != nil
	if ok {
		value = *r.page
	}
	return
}

// Search returns the value of the 'search' parameter.
//
// Search criteria.
//
// The syntax of this parameter is similar to the syntax of the _where_ clause
// of an SQL statement, but using the names of the attributes of the account
// instead of the names of the columns of a table. For example, in order to
// retrieve accounts with username starting with my:
//
// [source,sql]
// ----
// username like 'my%'
// ----
//
// If the parameter isn't provided, or if the value is empty, then all the
// items that the user has permission to see will be returned.
func (r *AccountsListServerRequest) Search() string {
	if r != nil && r.search != nil {
		return *r.search
	}
	return ""
}

// GetSearch returns the value of the 'search' parameter and
// a flag indicating if the parameter has a value.
//
// Search criteria.
//
// The syntax of this parameter is similar to the syntax of the _where_ clause
// of an SQL statement, but using the names of the attributes of the account
// instead of the names of the columns of a table. For example, in order to
// retrieve accounts with username starting with my:
//
// [source,sql]
// ----
// username like 'my%'
// ----
//
// If the parameter isn't provided, or if the value is empty, then all the
// items that the user has permission to see will be returned.
func (r *AccountsListServerRequest) GetSearch() (value string, ok bool) {
	ok = r != nil && r.search != nil
	if ok {
		value = *r.search
	}
	return
}

// Size returns the value of the 'size' parameter.
//
// Maximum number of items that will be contained in the returned page.
func (r *AccountsListServerRequest) Size() int {
	if r != nil && r.size != nil {
		return *r.size
	}
	return 0
}

// GetSize returns the value of the 'size' parameter and
// a flag indicating if the parameter has a value.
//
// Maximum number of items that will be contained in the returned page.
func (r *AccountsListServerRequest) GetSize() (value int, ok bool) {
	ok = r != nil && r.size != nil
	if ok {
		value = *r.size
	}
	return
}

// AccountsListServerResponse is the response for the 'list' method.
type AccountsListServerResponse struct {
	status int
	err    *errors.Error
	items  *AccountList
	page   *int
	size   *int
	total  *int
}

// Items sets the value of the 'items' parameter.
//
// Retrieved list of accounts.
func (r *AccountsListServerResponse) Items(value *AccountList) *AccountsListServerResponse {
	r.items = value
	return r
}

// Page sets the value of the 'page' parameter.
//
// Index of the requested page, where one corresponds to the first page.
func (r *AccountsListServerResponse) Page(value int) *AccountsListServerResponse {
	r.page = &value
	return r
}

// Size sets the value of the 'size' parameter.
//
// Maximum number of items that will be contained in the returned page.
func (r *AccountsListServerResponse) Size(value int) *AccountsListServerResponse {
	r.size = &value
	return r
}

// Total sets the value of the 'total' parameter.
//
// Total number of items of the collection that match the search criteria,
// regardless of the size of the page.
func (r *AccountsListServerResponse) Total(value int) *AccountsListServerResponse {
	r.total = &value
	return r
}

// Status sets the status code.
func (r *AccountsListServerResponse) Status(value int) *AccountsListServerResponse {
	r.status = value
	return r
}

// marshall is the method used internally to marshal responses for the
// 'list' method.
func (r *AccountsListServerResponse) marshal(writer io.Writer) error {
	var err error
	encoder := json.NewEncoder(writer)
	data := new(accountsListServerResponseData)
	data.Items, err = r.items.wrap()
	if err != nil {
		return err
	}
	data.Page = r.page
	data.Size = r.size
	data.Total = r.total
	err = encoder.Encode(data)
	return err
}

// accountsListServerResponseData is the structure used internally to write the request of the
// 'list' method.
type accountsListServerResponseData struct {
	Items accountListData "json:\"items,omitempty\""
	Page  *int            "json:\"page,omitempty\""
	Size  *int            "json:\"size,omitempty\""
	Total *int            "json:\"total,omitempty\""
}

// dispatchAccounts navigates the servers tree rooted at the given server
// till it finds one that matches the given set of path segments, and then invokes
// the corresponding server.
func dispatchAccounts(w http.ResponseWriter, r *http.Request, server AccountsServer, segments []string) {
	if len(segments) == 0 {
		switch r.Method {
		case "POST":
			adaptAccountsAddRequest(w, r, server)
		case "GET":
			adaptAccountsListRequest(w, r, server)
		default:
			errors.SendMethodNotAllowed(w, r)
			return
		}
	} else {
		switch segments[0] {
		default:
			target := server.Account(segments[0])
			if target == nil {
				errors.SendNotFound(w, r)
				return
			}
			dispatchAccount(w, r, target, segments[1:])
		}
	}
}

// readAccountsAddRequest reads the given HTTP requests and translates it
// into an object of type AccountsAddServerRequest.
func readAccountsAddRequest(r *http.Request) (*AccountsAddServerRequest, error) {
	var err error
	result := new(AccountsAddServerRequest)
	err = result.unmarshal(r.Body)
	if err != nil {
		return nil, err
	}
	return result, err
}

// writeAccountsAddResponse translates the given request object into an
// HTTP response.
func writeAccountsAddResponse(w http.ResponseWriter, r *AccountsAddServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptAccountsAddRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptAccountsAddRequest(w http.ResponseWriter, r *http.Request, server AccountsServer) {
	request, err := readAccountsAddRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(AccountsAddServerResponse)
	response.status = 201
	err = server.Add(r.Context(), request, response)
	if err != nil {
		glog.Errorf(
			"Can't process request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	err = writeAccountsAddResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}

// readAccountsListRequest reads the given HTTP requests and translates it
// into an object of type AccountsListServerRequest.
func readAccountsListRequest(r *http.Request) (*AccountsListServerRequest, error) {
	var err error
	result := new(AccountsListServerRequest)
	query := r.URL.Query()
	result.order, err = helpers.ParseString(query, "order")
	if err != nil {
		return nil, err
	}
	result.page, err = helpers.ParseInteger(query, "page")
	if err != nil {
		return nil, err
	}
	if result.page == nil {
		result.page = helpers.NewInteger(1)
	}
	result.search, err = helpers.ParseString(query, "search")
	if err != nil {
		return nil, err
	}
	result.size, err = helpers.ParseInteger(query, "size")
	if err != nil {
		return nil, err
	}
	if result.size == nil {
		result.size = helpers.NewInteger(100)
	}
	return result, err
}

// writeAccountsListResponse translates the given request object into an
// HTTP response.
func writeAccountsListResponse(w http.ResponseWriter, r *AccountsListServerResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.status)
	err := r.marshal(w)
	if err != nil {
		return err
	}
	return nil
}

// adaptAccountsListRequest translates the given HTTP request into a call to
// the corresponding method of the given server. Then it translates the
// results returned by that method into an HTTP response.
func adaptAccountsListRequest(w http.ResponseWriter, r *http.Request, server AccountsServer) {
	request, err := readAccountsListRequest(r)
	if err != nil {
		glog.Errorf(
			"Can't read request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	response := new(AccountsListServerResponse)
	response.status = 200
	err = server.List(r.Context(), request, response)
	if err != nil {
		glog.Errorf(
			"Can't process request for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		errors.SendInternalServerError(w, r)
		return
	}
	err = writeAccountsListResponse(w, response)
	if err != nil {
		glog.Errorf(
			"Can't write response for method '%s' and path '%s': %v",
			r.Method, r.URL.Path, err,
		)
		return
	}
}
