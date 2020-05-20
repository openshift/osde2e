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

package v1 // github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1

// CloudRegionKind is the name of the type used to represent objects
// of type 'cloud_region'.
const CloudRegionKind = "CloudRegion"

// CloudRegionLinkKind is the name of the type used to represent links
// to objects of type 'cloud_region'.
const CloudRegionLinkKind = "CloudRegionLink"

// CloudRegionNilKind is the name of the type used to nil references
// to objects of type 'cloud_region'.
const CloudRegionNilKind = "CloudRegionNil"

// CloudRegion represents the values of the 'cloud_region' type.
//
// Description of a region of a cloud provider.
type CloudRegion struct {
	id            *string
	href          *string
	link          bool
	cloudProvider *CloudProvider
	displayName   *string
	enabled       *bool
	name          *string
}

// Kind returns the name of the type of the object.
func (o *CloudRegion) Kind() string {
	if o == nil {
		return CloudRegionNilKind
	}
	if o.link {
		return CloudRegionLinkKind
	}
	return CloudRegionKind
}

// ID returns the identifier of the object.
func (o *CloudRegion) ID() string {
	if o != nil && o.id != nil {
		return *o.id
	}
	return ""
}

// GetID returns the identifier of the object and a flag indicating if the
// identifier has a value.
func (o *CloudRegion) GetID() (value string, ok bool) {
	ok = o != nil && o.id != nil
	if ok {
		value = *o.id
	}
	return
}

// Link returns true iif this is a link.
func (o *CloudRegion) Link() bool {
	return o != nil && o.link
}

// HREF returns the link to the object.
func (o *CloudRegion) HREF() string {
	if o != nil && o.href != nil {
		return *o.href
	}
	return ""
}

// GetHREF returns the link of the object and a flag indicating if the
// link has a value.
func (o *CloudRegion) GetHREF() (value string, ok bool) {
	ok = o != nil && o.href != nil
	if ok {
		value = *o.href
	}
	return
}

// Empty returns true if the object is empty, i.e. no attribute has a value.
func (o *CloudRegion) Empty() bool {
	return o == nil || (o.id == nil &&
		o.displayName == nil &&
		o.enabled == nil &&
		o.name == nil &&
		true)
}

// CloudProvider returns the value of the 'cloud_provider' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Link to the cloud provider that the region belongs to.
func (o *CloudRegion) CloudProvider() *CloudProvider {
	if o == nil {
		return nil
	}
	return o.cloudProvider
}

// GetCloudProvider returns the value of the 'cloud_provider' attribute and
// a flag indicating if the attribute has a value.
//
// Link to the cloud provider that the region belongs to.
func (o *CloudRegion) GetCloudProvider() (value *CloudProvider, ok bool) {
	ok = o != nil && o.cloudProvider != nil
	if ok {
		value = o.cloudProvider
	}
	return
}

// DisplayName returns the value of the 'display_name' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Name of the region for display purposes, for example `N. Virginia`.
func (o *CloudRegion) DisplayName() string {
	if o != nil && o.displayName != nil {
		return *o.displayName
	}
	return ""
}

// GetDisplayName returns the value of the 'display_name' attribute and
// a flag indicating if the attribute has a value.
//
// Name of the region for display purposes, for example `N. Virginia`.
func (o *CloudRegion) GetDisplayName() (value string, ok bool) {
	ok = o != nil && o.displayName != nil
	if ok {
		value = *o.displayName
	}
	return
}

// Enabled returns the value of the 'enabled' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Whether the region is enabled for deploying an OSD cluster.
func (o *CloudRegion) Enabled() bool {
	if o != nil && o.enabled != nil {
		return *o.enabled
	}
	return false
}

// GetEnabled returns the value of the 'enabled' attribute and
// a flag indicating if the attribute has a value.
//
// Whether the region is enabled for deploying an OSD cluster.
func (o *CloudRegion) GetEnabled() (value bool, ok bool) {
	ok = o != nil && o.enabled != nil
	if ok {
		value = *o.enabled
	}
	return
}

// Name returns the value of the 'name' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Human friendly identifier of the region, for example `us-east-1`.
//
// NOTE: Currently for all cloud providers and all regions `id` and `name` have exactly
// the same values.
func (o *CloudRegion) Name() string {
	if o != nil && o.name != nil {
		return *o.name
	}
	return ""
}

// GetName returns the value of the 'name' attribute and
// a flag indicating if the attribute has a value.
//
// Human friendly identifier of the region, for example `us-east-1`.
//
// NOTE: Currently for all cloud providers and all regions `id` and `name` have exactly
// the same values.
func (o *CloudRegion) GetName() (value string, ok bool) {
	ok = o != nil && o.name != nil
	if ok {
		value = *o.name
	}
	return
}

// CloudRegionListKind is the name of the type used to represent list of objects of
// type 'cloud_region'.
const CloudRegionListKind = "CloudRegionList"

// CloudRegionListLinkKind is the name of the type used to represent links to list
// of objects of type 'cloud_region'.
const CloudRegionListLinkKind = "CloudRegionListLink"

// CloudRegionNilKind is the name of the type used to nil lists of objects of
// type 'cloud_region'.
const CloudRegionListNilKind = "CloudRegionListNil"

// CloudRegionList is a list of values of the 'cloud_region' type.
type CloudRegionList struct {
	href  *string
	link  bool
	items []*CloudRegion
}

// Kind returns the name of the type of the object.
func (l *CloudRegionList) Kind() string {
	if l == nil {
		return CloudRegionListNilKind
	}
	if l.link {
		return CloudRegionListLinkKind
	}
	return CloudRegionListKind
}

// Link returns true iif this is a link.
func (l *CloudRegionList) Link() bool {
	return l != nil && l.link
}

// HREF returns the link to the list.
func (l *CloudRegionList) HREF() string {
	if l != nil && l.href != nil {
		return *l.href
	}
	return ""
}

// GetHREF returns the link of the list and a flag indicating if the
// link has a value.
func (l *CloudRegionList) GetHREF() (value string, ok bool) {
	ok = l != nil && l.href != nil
	if ok {
		value = *l.href
	}
	return
}

// Len returns the length of the list.
func (l *CloudRegionList) Len() int {
	if l == nil {
		return 0
	}
	return len(l.items)
}

// Empty returns true if the list is empty.
func (l *CloudRegionList) Empty() bool {
	return l == nil || len(l.items) == 0
}

// Get returns the item of the list with the given index. If there is no item with
// that index it returns nil.
func (l *CloudRegionList) Get(i int) *CloudRegion {
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
func (l *CloudRegionList) Slice() []*CloudRegion {
	var slice []*CloudRegion
	if l == nil {
		slice = make([]*CloudRegion, 0)
	} else {
		slice = make([]*CloudRegion, len(l.items))
		copy(slice, l.items)
	}
	return slice
}

// Each runs the given function for each item of the list, in order. If the function
// returns false the iteration stops, otherwise it continues till all the elements
// of the list have been processed.
func (l *CloudRegionList) Each(f func(item *CloudRegion) bool) {
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
func (l *CloudRegionList) Range(f func(index int, item *CloudRegion) bool) {
	if l == nil {
		return
	}
	for index, item := range l.items {
		if !f(index, item) {
			break
		}
	}
}
