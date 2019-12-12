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

// RoleBindingKind is the name of the type used to represent objects
// of type 'role_binding'.
const RoleBindingKind = "RoleBinding"

// RoleBindingLinkKind is the name of the type used to represent links
// to objects of type 'role_binding'.
const RoleBindingLinkKind = "RoleBindingLink"

// RoleBindingNilKind is the name of the type used to nil references
// to objects of type 'role_binding'.
const RoleBindingNilKind = "RoleBindingNil"

// RoleBinding represents the values of the 'role_binding' type.
//
//
type RoleBinding struct {
	id            *string
	href          *string
	link          bool
	account       *Account
	configManaged *bool
	organization  *Organization
	role          *Role
	subscription  *Subscription
	type_         *string
}

// Kind returns the name of the type of the object.
func (o *RoleBinding) Kind() string {
	if o == nil {
		return RoleBindingNilKind
	}
	if o.link {
		return RoleBindingLinkKind
	}
	return RoleBindingKind
}

// ID returns the identifier of the object.
func (o *RoleBinding) ID() string {
	if o != nil && o.id != nil {
		return *o.id
	}
	return ""
}

// GetID returns the identifier of the object and a flag indicating if the
// identifier has a value.
func (o *RoleBinding) GetID() (value string, ok bool) {
	ok = o != nil && o.id != nil
	if ok {
		value = *o.id
	}
	return
}

// Link returns true iif this is a link.
func (o *RoleBinding) Link() bool {
	return o != nil && o.link
}

// HREF returns the link to the object.
func (o *RoleBinding) HREF() string {
	if o != nil && o.href != nil {
		return *o.href
	}
	return ""
}

// GetHREF returns the link of the object and a flag indicating if the
// link has a value.
func (o *RoleBinding) GetHREF() (value string, ok bool) {
	ok = o != nil && o.href != nil
	if ok {
		value = *o.href
	}
	return
}

// Empty returns true if the object is empty, i.e. no attribute has a value.
func (o *RoleBinding) Empty() bool {
	return o == nil || (o.id == nil &&
		o.account == nil &&
		o.configManaged == nil &&
		o.organization == nil &&
		o.role == nil &&
		o.subscription == nil &&
		o.type_ == nil &&
		true)
}

// Account returns the value of the 'account' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *RoleBinding) Account() *Account {
	if o == nil {
		return nil
	}
	return o.account
}

// GetAccount returns the value of the 'account' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *RoleBinding) GetAccount() (value *Account, ok bool) {
	ok = o != nil && o.account != nil
	if ok {
		value = o.account
	}
	return
}

// ConfigManaged returns the value of the 'config_managed' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *RoleBinding) ConfigManaged() bool {
	if o != nil && o.configManaged != nil {
		return *o.configManaged
	}
	return false
}

// GetConfigManaged returns the value of the 'config_managed' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *RoleBinding) GetConfigManaged() (value bool, ok bool) {
	ok = o != nil && o.configManaged != nil
	if ok {
		value = *o.configManaged
	}
	return
}

// Organization returns the value of the 'organization' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *RoleBinding) Organization() *Organization {
	if o == nil {
		return nil
	}
	return o.organization
}

// GetOrganization returns the value of the 'organization' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *RoleBinding) GetOrganization() (value *Organization, ok bool) {
	ok = o != nil && o.organization != nil
	if ok {
		value = o.organization
	}
	return
}

// Role returns the value of the 'role' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *RoleBinding) Role() *Role {
	if o == nil {
		return nil
	}
	return o.role
}

// GetRole returns the value of the 'role' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *RoleBinding) GetRole() (value *Role, ok bool) {
	ok = o != nil && o.role != nil
	if ok {
		value = o.role
	}
	return
}

// Subscription returns the value of the 'subscription' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *RoleBinding) Subscription() *Subscription {
	if o == nil {
		return nil
	}
	return o.subscription
}

// GetSubscription returns the value of the 'subscription' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *RoleBinding) GetSubscription() (value *Subscription, ok bool) {
	ok = o != nil && o.subscription != nil
	if ok {
		value = o.subscription
	}
	return
}

// Type returns the value of the 'type' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *RoleBinding) Type() string {
	if o != nil && o.type_ != nil {
		return *o.type_
	}
	return ""
}

// GetType returns the value of the 'type' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *RoleBinding) GetType() (value string, ok bool) {
	ok = o != nil && o.type_ != nil
	if ok {
		value = *o.type_
	}
	return
}

// RoleBindingListKind is the name of the type used to represent list of
// objects of type 'role_binding'.
const RoleBindingListKind = "RoleBindingList"

// RoleBindingListLinkKind is the name of the type used to represent links
// to list of objects of type 'role_binding'.
const RoleBindingListLinkKind = "RoleBindingListLink"

// RoleBindingNilKind is the name of the type used to nil lists of
// objects of type 'role_binding'.
const RoleBindingListNilKind = "RoleBindingListNil"

// RoleBindingList is a list of values of the 'role_binding' type.
type RoleBindingList struct {
	href  *string
	link  bool
	items []*RoleBinding
}

// Kind returns the name of the type of the object.
func (l *RoleBindingList) Kind() string {
	if l == nil {
		return RoleBindingListNilKind
	}
	if l.link {
		return RoleBindingListLinkKind
	}
	return RoleBindingListKind
}

// Link returns true iif this is a link.
func (l *RoleBindingList) Link() bool {
	return l != nil && l.link
}

// HREF returns the link to the list.
func (l *RoleBindingList) HREF() string {
	if l != nil && l.href != nil {
		return *l.href
	}
	return ""
}

// GetHREF returns the link of the list and a flag indicating if the
// link has a value.
func (l *RoleBindingList) GetHREF() (value string, ok bool) {
	ok = l != nil && l.href != nil
	if ok {
		value = *l.href
	}
	return
}

// Len returns the length of the list.
func (l *RoleBindingList) Len() int {
	if l == nil {
		return 0
	}
	return len(l.items)
}

// Empty returns true if the list is empty.
func (l *RoleBindingList) Empty() bool {
	return l == nil || len(l.items) == 0
}

// Get returns the item of the list with the given index. If there is no item with
// that index it returns nil.
func (l *RoleBindingList) Get(i int) *RoleBinding {
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
func (l *RoleBindingList) Slice() []*RoleBinding {
	var slice []*RoleBinding
	if l == nil {
		slice = make([]*RoleBinding, 0)
	} else {
		slice = make([]*RoleBinding, len(l.items))
		copy(slice, l.items)
	}
	return slice
}

// Each runs the given function for each item of the list, in order. If the function
// returns false the iteration stops, otherwise it continues till all the elements
// of the list have been processed.
func (l *RoleBindingList) Each(f func(item *RoleBinding) bool) {
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
func (l *RoleBindingList) Range(f func(index int, item *RoleBinding) bool) {
	if l == nil {
		return
	}
	for index, item := range l.items {
		if !f(index, item) {
			break
		}
	}
}
