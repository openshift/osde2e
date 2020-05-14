/*
Copyright (c) 2020 Red Hat, Inc.

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

// ClusterAuthorizationRequest represents the values of the 'cluster_authorization_request' type.
//
//
type ClusterAuthorizationRequest struct {
	byoc              *bool
	accountUsername   *string
	availabilityZone  *string
	clusterID         *string
	disconnected      *bool
	displayName       *string
	externalClusterID *string
	managed           *bool
	reserve           *bool
	resources         []*ReservedResource
}

// Empty returns true if the object is empty, i.e. no attribute has a value.
func (o *ClusterAuthorizationRequest) Empty() bool {
	return o == nil || (o.byoc == nil &&
		o.accountUsername == nil &&
		o.availabilityZone == nil &&
		o.clusterID == nil &&
		o.disconnected == nil &&
		o.displayName == nil &&
		o.externalClusterID == nil &&
		o.managed == nil &&
		o.reserve == nil &&
		len(o.resources) == 0 &&
		true)
}

// BYOC returns the value of the 'BYOC' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *ClusterAuthorizationRequest) BYOC() bool {
	if o != nil && o.byoc != nil {
		return *o.byoc
	}
	return false
}

// GetBYOC returns the value of the 'BYOC' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *ClusterAuthorizationRequest) GetBYOC() (value bool, ok bool) {
	ok = o != nil && o.byoc != nil
	if ok {
		value = *o.byoc
	}
	return
}

// AccountUsername returns the value of the 'account_username' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *ClusterAuthorizationRequest) AccountUsername() string {
	if o != nil && o.accountUsername != nil {
		return *o.accountUsername
	}
	return ""
}

// GetAccountUsername returns the value of the 'account_username' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *ClusterAuthorizationRequest) GetAccountUsername() (value string, ok bool) {
	ok = o != nil && o.accountUsername != nil
	if ok {
		value = *o.accountUsername
	}
	return
}

// AvailabilityZone returns the value of the 'availability_zone' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *ClusterAuthorizationRequest) AvailabilityZone() string {
	if o != nil && o.availabilityZone != nil {
		return *o.availabilityZone
	}
	return ""
}

// GetAvailabilityZone returns the value of the 'availability_zone' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *ClusterAuthorizationRequest) GetAvailabilityZone() (value string, ok bool) {
	ok = o != nil && o.availabilityZone != nil
	if ok {
		value = *o.availabilityZone
	}
	return
}

// ClusterID returns the value of the 'cluster_ID' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *ClusterAuthorizationRequest) ClusterID() string {
	if o != nil && o.clusterID != nil {
		return *o.clusterID
	}
	return ""
}

// GetClusterID returns the value of the 'cluster_ID' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *ClusterAuthorizationRequest) GetClusterID() (value string, ok bool) {
	ok = o != nil && o.clusterID != nil
	if ok {
		value = *o.clusterID
	}
	return
}

// Disconnected returns the value of the 'disconnected' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *ClusterAuthorizationRequest) Disconnected() bool {
	if o != nil && o.disconnected != nil {
		return *o.disconnected
	}
	return false
}

// GetDisconnected returns the value of the 'disconnected' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *ClusterAuthorizationRequest) GetDisconnected() (value bool, ok bool) {
	ok = o != nil && o.disconnected != nil
	if ok {
		value = *o.disconnected
	}
	return
}

// DisplayName returns the value of the 'display_name' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *ClusterAuthorizationRequest) DisplayName() string {
	if o != nil && o.displayName != nil {
		return *o.displayName
	}
	return ""
}

// GetDisplayName returns the value of the 'display_name' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *ClusterAuthorizationRequest) GetDisplayName() (value string, ok bool) {
	ok = o != nil && o.displayName != nil
	if ok {
		value = *o.displayName
	}
	return
}

// ExternalClusterID returns the value of the 'external_cluster_ID' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *ClusterAuthorizationRequest) ExternalClusterID() string {
	if o != nil && o.externalClusterID != nil {
		return *o.externalClusterID
	}
	return ""
}

// GetExternalClusterID returns the value of the 'external_cluster_ID' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *ClusterAuthorizationRequest) GetExternalClusterID() (value string, ok bool) {
	ok = o != nil && o.externalClusterID != nil
	if ok {
		value = *o.externalClusterID
	}
	return
}

// Managed returns the value of the 'managed' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *ClusterAuthorizationRequest) Managed() bool {
	if o != nil && o.managed != nil {
		return *o.managed
	}
	return false
}

// GetManaged returns the value of the 'managed' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *ClusterAuthorizationRequest) GetManaged() (value bool, ok bool) {
	ok = o != nil && o.managed != nil
	if ok {
		value = *o.managed
	}
	return
}

// Reserve returns the value of the 'reserve' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *ClusterAuthorizationRequest) Reserve() bool {
	if o != nil && o.reserve != nil {
		return *o.reserve
	}
	return false
}

// GetReserve returns the value of the 'reserve' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *ClusterAuthorizationRequest) GetReserve() (value bool, ok bool) {
	ok = o != nil && o.reserve != nil
	if ok {
		value = *o.reserve
	}
	return
}

// Resources returns the value of the 'resources' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *ClusterAuthorizationRequest) Resources() []*ReservedResource {
	if o == nil {
		return nil
	}
	return o.resources
}

// GetResources returns the value of the 'resources' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *ClusterAuthorizationRequest) GetResources() (value []*ReservedResource, ok bool) {
	ok = o != nil && o.resources != nil
	if ok {
		value = o.resources
	}
	return
}

// ClusterAuthorizationRequestListKind is the name of the type used to represent list of objects of
// type 'cluster_authorization_request'.
const ClusterAuthorizationRequestListKind = "ClusterAuthorizationRequestList"

// ClusterAuthorizationRequestListLinkKind is the name of the type used to represent links to list
// of objects of type 'cluster_authorization_request'.
const ClusterAuthorizationRequestListLinkKind = "ClusterAuthorizationRequestListLink"

// ClusterAuthorizationRequestNilKind is the name of the type used to nil lists of objects of
// type 'cluster_authorization_request'.
const ClusterAuthorizationRequestListNilKind = "ClusterAuthorizationRequestListNil"

// ClusterAuthorizationRequestList is a list of values of the 'cluster_authorization_request' type.
type ClusterAuthorizationRequestList struct {
	href  *string
	link  bool
	items []*ClusterAuthorizationRequest
}

// Len returns the length of the list.
func (l *ClusterAuthorizationRequestList) Len() int {
	if l == nil {
		return 0
	}
	return len(l.items)
}

// Empty returns true if the list is empty.
func (l *ClusterAuthorizationRequestList) Empty() bool {
	return l == nil || len(l.items) == 0
}

// Get returns the item of the list with the given index. If there is no item with
// that index it returns nil.
func (l *ClusterAuthorizationRequestList) Get(i int) *ClusterAuthorizationRequest {
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
func (l *ClusterAuthorizationRequestList) Slice() []*ClusterAuthorizationRequest {
	var slice []*ClusterAuthorizationRequest
	if l == nil {
		slice = make([]*ClusterAuthorizationRequest, 0)
	} else {
		slice = make([]*ClusterAuthorizationRequest, len(l.items))
		copy(slice, l.items)
	}
	return slice
}

// Each runs the given function for each item of the list, in order. If the function
// returns false the iteration stops, otherwise it continues till all the elements
// of the list have been processed.
func (l *ClusterAuthorizationRequestList) Each(f func(item *ClusterAuthorizationRequest) bool) {
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
func (l *ClusterAuthorizationRequestList) Range(f func(index int, item *ClusterAuthorizationRequest) bool) {
	if l == nil {
		return
	}
	for index, item := range l.items {
		if !f(index, item) {
			break
		}
	}
}
