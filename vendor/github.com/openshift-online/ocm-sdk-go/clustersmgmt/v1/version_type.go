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

// VersionKind is the name of the type used to represent objects
// of type 'version'.
const VersionKind = "Version"

// VersionLinkKind is the name of the type used to represent links
// to objects of type 'version'.
const VersionLinkKind = "VersionLink"

// VersionNilKind is the name of the type used to nil references
// to objects of type 'version'.
const VersionNilKind = "VersionNil"

// Version represents the values of the 'version' type.
//
// Representation of an _OpenShift_ version.
type Version struct {
	id                *string
	href              *string
	link              bool
	moaEnabled        *bool
	availableUpgrades []string
	channelGroup      *string
	default_          *bool
	enabled           *bool
}

// Kind returns the name of the type of the object.
func (o *Version) Kind() string {
	if o == nil {
		return VersionNilKind
	}
	if o.link {
		return VersionLinkKind
	}
	return VersionKind
}

// ID returns the identifier of the object.
func (o *Version) ID() string {
	if o != nil && o.id != nil {
		return *o.id
	}
	return ""
}

// GetID returns the identifier of the object and a flag indicating if the
// identifier has a value.
func (o *Version) GetID() (value string, ok bool) {
	ok = o != nil && o.id != nil
	if ok {
		value = *o.id
	}
	return
}

// Link returns true iif this is a link.
func (o *Version) Link() bool {
	return o != nil && o.link
}

// HREF returns the link to the object.
func (o *Version) HREF() string {
	if o != nil && o.href != nil {
		return *o.href
	}
	return ""
}

// GetHREF returns the link of the object and a flag indicating if the
// link has a value.
func (o *Version) GetHREF() (value string, ok bool) {
	ok = o != nil && o.href != nil
	if ok {
		value = *o.href
	}
	return
}

// Empty returns true if the object is empty, i.e. no attribute has a value.
func (o *Version) Empty() bool {
	return o == nil || (o.id == nil &&
		o.moaEnabled == nil &&
		len(o.availableUpgrades) == 0 &&
		o.channelGroup == nil &&
		o.default_ == nil &&
		o.enabled == nil &&
		true)
}

// MOAEnabled returns the value of the 'MOA_enabled' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// MOAEnabled indicates whether this version can be used to create MOA clusters.
func (o *Version) MOAEnabled() bool {
	if o != nil && o.moaEnabled != nil {
		return *o.moaEnabled
	}
	return false
}

// GetMOAEnabled returns the value of the 'MOA_enabled' attribute and
// a flag indicating if the attribute has a value.
//
// MOAEnabled indicates whether this version can be used to create MOA clusters.
func (o *Version) GetMOAEnabled() (value bool, ok bool) {
	ok = o != nil && o.moaEnabled != nil
	if ok {
		value = *o.moaEnabled
	}
	return
}

// AvailableUpgrades returns the value of the 'available_upgrades' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// AvailableUpgrades is the list of versions this version can be upgraded to.
func (o *Version) AvailableUpgrades() []string {
	if o == nil {
		return nil
	}
	return o.availableUpgrades
}

// GetAvailableUpgrades returns the value of the 'available_upgrades' attribute and
// a flag indicating if the attribute has a value.
//
// AvailableUpgrades is the list of versions this version can be upgraded to.
func (o *Version) GetAvailableUpgrades() (value []string, ok bool) {
	ok = o != nil && o.availableUpgrades != nil
	if ok {
		value = o.availableUpgrades
	}
	return
}

// ChannelGroup returns the value of the 'channel_group' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// ChannelGroup is the name of the group where this image belongs.
// ChannelGroup is a mechanism to partition the images to different groups,
// each image belongs to only a single group.
func (o *Version) ChannelGroup() string {
	if o != nil && o.channelGroup != nil {
		return *o.channelGroup
	}
	return ""
}

// GetChannelGroup returns the value of the 'channel_group' attribute and
// a flag indicating if the attribute has a value.
//
// ChannelGroup is the name of the group where this image belongs.
// ChannelGroup is a mechanism to partition the images to different groups,
// each image belongs to only a single group.
func (o *Version) GetChannelGroup() (value string, ok bool) {
	ok = o != nil && o.channelGroup != nil
	if ok {
		value = *o.channelGroup
	}
	return
}

// Default returns the value of the 'default' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Indicates if this should be selected as the default version when a cluster is created
// without specifying explicitly the version.
func (o *Version) Default() bool {
	if o != nil && o.default_ != nil {
		return *o.default_
	}
	return false
}

// GetDefault returns the value of the 'default' attribute and
// a flag indicating if the attribute has a value.
//
// Indicates if this should be selected as the default version when a cluster is created
// without specifying explicitly the version.
func (o *Version) GetDefault() (value bool, ok bool) {
	ok = o != nil && o.default_ != nil
	if ok {
		value = *o.default_
	}
	return
}

// Enabled returns the value of the 'enabled' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Indicates if this version can be used to create clusters.
func (o *Version) Enabled() bool {
	if o != nil && o.enabled != nil {
		return *o.enabled
	}
	return false
}

// GetEnabled returns the value of the 'enabled' attribute and
// a flag indicating if the attribute has a value.
//
// Indicates if this version can be used to create clusters.
func (o *Version) GetEnabled() (value bool, ok bool) {
	ok = o != nil && o.enabled != nil
	if ok {
		value = *o.enabled
	}
	return
}

// VersionListKind is the name of the type used to represent list of objects of
// type 'version'.
const VersionListKind = "VersionList"

// VersionListLinkKind is the name of the type used to represent links to list
// of objects of type 'version'.
const VersionListLinkKind = "VersionListLink"

// VersionNilKind is the name of the type used to nil lists of objects of
// type 'version'.
const VersionListNilKind = "VersionListNil"

// VersionList is a list of values of the 'version' type.
type VersionList struct {
	href  *string
	link  bool
	items []*Version
}

// Kind returns the name of the type of the object.
func (l *VersionList) Kind() string {
	if l == nil {
		return VersionListNilKind
	}
	if l.link {
		return VersionListLinkKind
	}
	return VersionListKind
}

// Link returns true iif this is a link.
func (l *VersionList) Link() bool {
	return l != nil && l.link
}

// HREF returns the link to the list.
func (l *VersionList) HREF() string {
	if l != nil && l.href != nil {
		return *l.href
	}
	return ""
}

// GetHREF returns the link of the list and a flag indicating if the
// link has a value.
func (l *VersionList) GetHREF() (value string, ok bool) {
	ok = l != nil && l.href != nil
	if ok {
		value = *l.href
	}
	return
}

// Len returns the length of the list.
func (l *VersionList) Len() int {
	if l == nil {
		return 0
	}
	return len(l.items)
}

// Empty returns true if the list is empty.
func (l *VersionList) Empty() bool {
	return l == nil || len(l.items) == 0
}

// Get returns the item of the list with the given index. If there is no item with
// that index it returns nil.
func (l *VersionList) Get(i int) *Version {
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
func (l *VersionList) Slice() []*Version {
	var slice []*Version
	if l == nil {
		slice = make([]*Version, 0)
	} else {
		slice = make([]*Version, len(l.items))
		copy(slice, l.items)
	}
	return slice
}

// Each runs the given function for each item of the list, in order. If the function
// returns false the iteration stops, otherwise it continues till all the elements
// of the list have been processed.
func (l *VersionList) Each(f func(item *Version) bool) {
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
func (l *VersionList) Range(f func(index int, item *Version) bool) {
	if l == nil {
		return
	}
	for index, item := range l.items {
		if !f(index, item) {
			break
		}
	}
}
