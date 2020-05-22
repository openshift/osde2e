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

// AddOnKind is the name of the type used to represent objects
// of type 'add_on'.
const AddOnKind = "AddOn"

// AddOnLinkKind is the name of the type used to represent links
// to objects of type 'add_on'.
const AddOnLinkKind = "AddOnLink"

// AddOnNilKind is the name of the type used to nil references
// to objects of type 'add_on'.
const AddOnNilKind = "AddOnNil"

// AddOn represents the values of the 'add_on' type.
//
// Representation of an add-on that can be installed in a cluster.
type AddOn struct {
	id              *string
	href            *string
	link            bool
	description     *string
	docsLink        *string
	enabled         *bool
	icon            *string
	installMode     *AddOnInstallMode
	label           *string
	name            *string
	operatorName    *string
	resourceCost    *float64
	resourceName    *string
	targetNamespace *string
}

// Kind returns the name of the type of the object.
func (o *AddOn) Kind() string {
	if o == nil {
		return AddOnNilKind
	}
	if o.link {
		return AddOnLinkKind
	}
	return AddOnKind
}

// ID returns the identifier of the object.
func (o *AddOn) ID() string {
	if o != nil && o.id != nil {
		return *o.id
	}
	return ""
}

// GetID returns the identifier of the object and a flag indicating if the
// identifier has a value.
func (o *AddOn) GetID() (value string, ok bool) {
	ok = o != nil && o.id != nil
	if ok {
		value = *o.id
	}
	return
}

// Link returns true iif this is a link.
func (o *AddOn) Link() bool {
	return o != nil && o.link
}

// HREF returns the link to the object.
func (o *AddOn) HREF() string {
	if o != nil && o.href != nil {
		return *o.href
	}
	return ""
}

// GetHREF returns the link of the object and a flag indicating if the
// link has a value.
func (o *AddOn) GetHREF() (value string, ok bool) {
	ok = o != nil && o.href != nil
	if ok {
		value = *o.href
	}
	return
}

// Empty returns true if the object is empty, i.e. no attribute has a value.
func (o *AddOn) Empty() bool {
	return o == nil || (o.id == nil &&
		o.description == nil &&
		o.docsLink == nil &&
		o.enabled == nil &&
		o.icon == nil &&
		o.installMode == nil &&
		o.label == nil &&
		o.name == nil &&
		o.operatorName == nil &&
		o.resourceCost == nil &&
		o.resourceName == nil &&
		o.targetNamespace == nil &&
		true)
}

// Description returns the value of the 'description' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Description of the add-on.
func (o *AddOn) Description() string {
	if o != nil && o.description != nil {
		return *o.description
	}
	return ""
}

// GetDescription returns the value of the 'description' attribute and
// a flag indicating if the attribute has a value.
//
// Description of the add-on.
func (o *AddOn) GetDescription() (value string, ok bool) {
	ok = o != nil && o.description != nil
	if ok {
		value = *o.description
	}
	return
}

// DocsLink returns the value of the 'docs_link' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Link to documentation about the add-on.
func (o *AddOn) DocsLink() string {
	if o != nil && o.docsLink != nil {
		return *o.docsLink
	}
	return ""
}

// GetDocsLink returns the value of the 'docs_link' attribute and
// a flag indicating if the attribute has a value.
//
// Link to documentation about the add-on.
func (o *AddOn) GetDocsLink() (value string, ok bool) {
	ok = o != nil && o.docsLink != nil
	if ok {
		value = *o.docsLink
	}
	return
}

// Enabled returns the value of the 'enabled' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Indicates if this add-on can be added to clusters.
func (o *AddOn) Enabled() bool {
	if o != nil && o.enabled != nil {
		return *o.enabled
	}
	return false
}

// GetEnabled returns the value of the 'enabled' attribute and
// a flag indicating if the attribute has a value.
//
// Indicates if this add-on can be added to clusters.
func (o *AddOn) GetEnabled() (value bool, ok bool) {
	ok = o != nil && o.enabled != nil
	if ok {
		value = *o.enabled
	}
	return
}

// Icon returns the value of the 'icon' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Base64-encoded icon representing an add-on. The icon should be in PNG format.
func (o *AddOn) Icon() string {
	if o != nil && o.icon != nil {
		return *o.icon
	}
	return ""
}

// GetIcon returns the value of the 'icon' attribute and
// a flag indicating if the attribute has a value.
//
// Base64-encoded icon representing an add-on. The icon should be in PNG format.
func (o *AddOn) GetIcon() (value string, ok bool) {
	ok = o != nil && o.icon != nil
	if ok {
		value = *o.icon
	}
	return
}

// InstallMode returns the value of the 'install_mode' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// The mode in which the addon is deployed.
func (o *AddOn) InstallMode() AddOnInstallMode {
	if o != nil && o.installMode != nil {
		return *o.installMode
	}
	return AddOnInstallMode("")
}

// GetInstallMode returns the value of the 'install_mode' attribute and
// a flag indicating if the attribute has a value.
//
// The mode in which the addon is deployed.
func (o *AddOn) GetInstallMode() (value AddOnInstallMode, ok bool) {
	ok = o != nil && o.installMode != nil
	if ok {
		value = *o.installMode
	}
	return
}

// Label returns the value of the 'label' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Label used to attach to a cluster deployment when add-on is installed.
func (o *AddOn) Label() string {
	if o != nil && o.label != nil {
		return *o.label
	}
	return ""
}

// GetLabel returns the value of the 'label' attribute and
// a flag indicating if the attribute has a value.
//
// Label used to attach to a cluster deployment when add-on is installed.
func (o *AddOn) GetLabel() (value string, ok bool) {
	ok = o != nil && o.label != nil
	if ok {
		value = *o.label
	}
	return
}

// Name returns the value of the 'name' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Name of the add-on.
func (o *AddOn) Name() string {
	if o != nil && o.name != nil {
		return *o.name
	}
	return ""
}

// GetName returns the value of the 'name' attribute and
// a flag indicating if the attribute has a value.
//
// Name of the add-on.
func (o *AddOn) GetName() (value string, ok bool) {
	ok = o != nil && o.name != nil
	if ok {
		value = *o.name
	}
	return
}

// OperatorName returns the value of the 'operator_name' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// The name of the operator installed by this add-on.
func (o *AddOn) OperatorName() string {
	if o != nil && o.operatorName != nil {
		return *o.operatorName
	}
	return ""
}

// GetOperatorName returns the value of the 'operator_name' attribute and
// a flag indicating if the attribute has a value.
//
// The name of the operator installed by this add-on.
func (o *AddOn) GetOperatorName() (value string, ok bool) {
	ok = o != nil && o.operatorName != nil
	if ok {
		value = *o.operatorName
	}
	return
}

// ResourceCost returns the value of the 'resource_cost' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Used to determine how many units of quota an add-on consumes per resource name.
func (o *AddOn) ResourceCost() float64 {
	if o != nil && o.resourceCost != nil {
		return *o.resourceCost
	}
	return 0.0
}

// GetResourceCost returns the value of the 'resource_cost' attribute and
// a flag indicating if the attribute has a value.
//
// Used to determine how many units of quota an add-on consumes per resource name.
func (o *AddOn) GetResourceCost() (value float64, ok bool) {
	ok = o != nil && o.resourceCost != nil
	if ok {
		value = *o.resourceCost
	}
	return
}

// ResourceName returns the value of the 'resource_name' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Used to determine from where to reserve quota for this add-on.
func (o *AddOn) ResourceName() string {
	if o != nil && o.resourceName != nil {
		return *o.resourceName
	}
	return ""
}

// GetResourceName returns the value of the 'resource_name' attribute and
// a flag indicating if the attribute has a value.
//
// Used to determine from where to reserve quota for this add-on.
func (o *AddOn) GetResourceName() (value string, ok bool) {
	ok = o != nil && o.resourceName != nil
	if ok {
		value = *o.resourceName
	}
	return
}

// TargetNamespace returns the value of the 'target_namespace' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// The namespace in which the addon CRD exists.
func (o *AddOn) TargetNamespace() string {
	if o != nil && o.targetNamespace != nil {
		return *o.targetNamespace
	}
	return ""
}

// GetTargetNamespace returns the value of the 'target_namespace' attribute and
// a flag indicating if the attribute has a value.
//
// The namespace in which the addon CRD exists.
func (o *AddOn) GetTargetNamespace() (value string, ok bool) {
	ok = o != nil && o.targetNamespace != nil
	if ok {
		value = *o.targetNamespace
	}
	return
}

// AddOnListKind is the name of the type used to represent list of objects of
// type 'add_on'.
const AddOnListKind = "AddOnList"

// AddOnListLinkKind is the name of the type used to represent links to list
// of objects of type 'add_on'.
const AddOnListLinkKind = "AddOnListLink"

// AddOnNilKind is the name of the type used to nil lists of objects of
// type 'add_on'.
const AddOnListNilKind = "AddOnListNil"

// AddOnList is a list of values of the 'add_on' type.
type AddOnList struct {
	href  *string
	link  bool
	items []*AddOn
}

// Kind returns the name of the type of the object.
func (l *AddOnList) Kind() string {
	if l == nil {
		return AddOnListNilKind
	}
	if l.link {
		return AddOnListLinkKind
	}
	return AddOnListKind
}

// Link returns true iif this is a link.
func (l *AddOnList) Link() bool {
	return l != nil && l.link
}

// HREF returns the link to the list.
func (l *AddOnList) HREF() string {
	if l != nil && l.href != nil {
		return *l.href
	}
	return ""
}

// GetHREF returns the link of the list and a flag indicating if the
// link has a value.
func (l *AddOnList) GetHREF() (value string, ok bool) {
	ok = l != nil && l.href != nil
	if ok {
		value = *l.href
	}
	return
}

// Len returns the length of the list.
func (l *AddOnList) Len() int {
	if l == nil {
		return 0
	}
	return len(l.items)
}

// Empty returns true if the list is empty.
func (l *AddOnList) Empty() bool {
	return l == nil || len(l.items) == 0
}

// Get returns the item of the list with the given index. If there is no item with
// that index it returns nil.
func (l *AddOnList) Get(i int) *AddOn {
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
func (l *AddOnList) Slice() []*AddOn {
	var slice []*AddOn
	if l == nil {
		slice = make([]*AddOn, 0)
	} else {
		slice = make([]*AddOn, len(l.items))
		copy(slice, l.items)
	}
	return slice
}

// Each runs the given function for each item of the list, in order. If the function
// returns false the iteration stops, otherwise it continues till all the elements
// of the list have been processed.
func (l *AddOnList) Each(f func(item *AddOn) bool) {
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
func (l *AddOnList) Range(f func(index int, item *AddOn) bool) {
	if l == nil {
		return
	}
	for index, item := range l.items {
		if !f(index, item) {
			break
		}
	}
}
