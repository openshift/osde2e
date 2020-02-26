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

// OpenIDURLs represents the values of the 'open_IDURLs' type.
//
// _OpenID_ identity provider URLs.
type OpenIDURLs struct {
	authorize *string
	token     *string
	userInfo  *string
}

// Empty returns true if the object is empty, i.e. no attribute has a value.
func (o *OpenIDURLs) Empty() bool {
	return o == nil || (o.authorize == nil &&
		o.token == nil &&
		o.userInfo == nil &&
		true)
}

// Authorize returns the value of the 'authorize' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Authorization endpoint described in the _OpenID_ specification. Must use HTTPS.
func (o *OpenIDURLs) Authorize() string {
	if o != nil && o.authorize != nil {
		return *o.authorize
	}
	return ""
}

// GetAuthorize returns the value of the 'authorize' attribute and
// a flag indicating if the attribute has a value.
//
// Authorization endpoint described in the _OpenID_ specification. Must use HTTPS.
func (o *OpenIDURLs) GetAuthorize() (value string, ok bool) {
	ok = o != nil && o.authorize != nil
	if ok {
		value = *o.authorize
	}
	return
}

// Token returns the value of the 'token' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Token endpoint described in the _OpenID_ specification. Must use HTTPS.
func (o *OpenIDURLs) Token() string {
	if o != nil && o.token != nil {
		return *o.token
	}
	return ""
}

// GetToken returns the value of the 'token' attribute and
// a flag indicating if the attribute has a value.
//
// Token endpoint described in the _OpenID_ specification. Must use HTTPS.
func (o *OpenIDURLs) GetToken() (value string, ok bool) {
	ok = o != nil && o.token != nil
	if ok {
		value = *o.token
	}
	return
}

// UserInfo returns the value of the 'user_info' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// User information endpoint described in the _OpenID_ specification. Must use HTTPS.
func (o *OpenIDURLs) UserInfo() string {
	if o != nil && o.userInfo != nil {
		return *o.userInfo
	}
	return ""
}

// GetUserInfo returns the value of the 'user_info' attribute and
// a flag indicating if the attribute has a value.
//
// User information endpoint described in the _OpenID_ specification. Must use HTTPS.
func (o *OpenIDURLs) GetUserInfo() (value string, ok bool) {
	ok = o != nil && o.userInfo != nil
	if ok {
		value = *o.userInfo
	}
	return
}

// OpenIDURLsListKind is the name of the type used to represent list of objects of
// type 'open_IDURLs'.
const OpenIDURLsListKind = "OpenIDURLsList"

// OpenIDURLsListLinkKind is the name of the type used to represent links to list
// of objects of type 'open_IDURLs'.
const OpenIDURLsListLinkKind = "OpenIDURLsListLink"

// OpenIDURLsNilKind is the name of the type used to nil lists of objects of
// type 'open_IDURLs'.
const OpenIDURLsListNilKind = "OpenIDURLsListNil"

// OpenIDURLsList is a list of values of the 'open_IDURLs' type.
type OpenIDURLsList struct {
	href  *string
	link  bool
	items []*OpenIDURLs
}

// Len returns the length of the list.
func (l *OpenIDURLsList) Len() int {
	if l == nil {
		return 0
	}
	return len(l.items)
}

// Empty returns true if the list is empty.
func (l *OpenIDURLsList) Empty() bool {
	return l == nil || len(l.items) == 0
}

// Get returns the item of the list with the given index. If there is no item with
// that index it returns nil.
func (l *OpenIDURLsList) Get(i int) *OpenIDURLs {
	if l == nil || i < 0 || i >= len(l.items) {
		return nil
	}
	return l.items[i]
}

// Slice returns an slice containing the items of the list. The returned slice is a
// copy of the one used internally, so it can be modified without affecting the
// internal representation.
//
// If you don't need to modify the returned slice consider using the Each or Range
// functions, as they don't need to allocate a new slice.
func (l *OpenIDURLsList) Slice() []*OpenIDURLs {
	var slice []*OpenIDURLs
	if l == nil {
		slice = make([]*OpenIDURLs, 0)
	} else {
		slice = make([]*OpenIDURLs, len(l.items))
		copy(slice, l.items)
	}
	return slice
}

// Each runs the given function for each item of the list, in order. If the function
// returns false the iteration stops, otherwise it continues till all the elements
// of the list have been processed.
func (l *OpenIDURLsList) Each(f func(item *OpenIDURLs) bool) {
	if l == nil {
		return
	}
	for _, item := range l.items {
		if !f(item) {
			break
		}
	}
}

// Range runs the given function for each index and item of the list, in order. If
// the function returns false the iteration stops, otherwise it continues till all
// the elements of the list have been processed.
func (l *OpenIDURLsList) Range(f func(index int, item *OpenIDURLs) bool) {
	if l == nil {
		return
	}
	for index, item := range l.items {
		if !f(index, item) {
			break
		}
	}
}
