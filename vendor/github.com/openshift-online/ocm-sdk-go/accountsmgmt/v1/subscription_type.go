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
	time "time"
)

// SubscriptionKind is the name of the type used to represent objects
// of type 'subscription'.
const SubscriptionKind = "Subscription"

// SubscriptionLinkKind is the name of the type used to represent links
// to objects of type 'subscription'.
const SubscriptionLinkKind = "SubscriptionLink"

// SubscriptionNilKind is the name of the type used to nil references
// to objects of type 'subscription'.
const SubscriptionNilKind = "SubscriptionNil"

// Subscription represents the values of the 'subscription' type.
//
//
type Subscription struct {
	id                 *string
	href               *string
	link               bool
	clusterID          *string
	createdAt          *time.Time
	creator            *Account
	displayName        *string
	externalClusterID  *string
	lastTelemetryDate  *time.Time
	organizationID     *string
	plan               *Plan
	registryCredential *RegistryCredential
	updatedAt          *time.Time
}

// Kind returns the name of the type of the object.
func (o *Subscription) Kind() string {
	if o == nil {
		return SubscriptionNilKind
	}
	if o.link {
		return SubscriptionLinkKind
	}
	return SubscriptionKind
}

// ID returns the identifier of the object.
func (o *Subscription) ID() string {
	if o != nil && o.id != nil {
		return *o.id
	}
	return ""
}

// GetID returns the identifier of the object and a flag indicating if the
// identifier has a value.
func (o *Subscription) GetID() (value string, ok bool) {
	ok = o != nil && o.id != nil
	if ok {
		value = *o.id
	}
	return
}

// Link returns true iif this is a link.
func (o *Subscription) Link() bool {
	return o != nil && o.link
}

// HREF returns the link to the object.
func (o *Subscription) HREF() string {
	if o != nil && o.href != nil {
		return *o.href
	}
	return ""
}

// GetHREF returns the link of the object and a flag indicating if the
// link has a value.
func (o *Subscription) GetHREF() (value string, ok bool) {
	ok = o != nil && o.href != nil
	if ok {
		value = *o.href
	}
	return
}

// Empty returns true if the object is empty, i.e. no attribute has a value.
func (o *Subscription) Empty() bool {
	return o == nil || (o.id == nil &&
		o.clusterID == nil &&
		o.createdAt == nil &&
		o.creator == nil &&
		o.displayName == nil &&
		o.externalClusterID == nil &&
		o.lastTelemetryDate == nil &&
		o.organizationID == nil &&
		o.plan == nil &&
		o.registryCredential == nil &&
		o.updatedAt == nil &&
		true)
}

// ClusterID returns the value of the 'cluster_ID' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *Subscription) ClusterID() string {
	if o != nil && o.clusterID != nil {
		return *o.clusterID
	}
	return ""
}

// GetClusterID returns the value of the 'cluster_ID' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *Subscription) GetClusterID() (value string, ok bool) {
	ok = o != nil && o.clusterID != nil
	if ok {
		value = *o.clusterID
	}
	return
}

// CreatedAt returns the value of the 'created_at' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *Subscription) CreatedAt() time.Time {
	if o != nil && o.createdAt != nil {
		return *o.createdAt
	}
	return time.Time{}
}

// GetCreatedAt returns the value of the 'created_at' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *Subscription) GetCreatedAt() (value time.Time, ok bool) {
	ok = o != nil && o.createdAt != nil
	if ok {
		value = *o.createdAt
	}
	return
}

// Creator returns the value of the 'creator' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Link to the account that created the subscription.
func (o *Subscription) Creator() *Account {
	if o == nil {
		return nil
	}
	return o.creator
}

// GetCreator returns the value of the 'creator' attribute and
// a flag indicating if the attribute has a value.
//
// Link to the account that created the subscription.
func (o *Subscription) GetCreator() (value *Account, ok bool) {
	ok = o != nil && o.creator != nil
	if ok {
		value = o.creator
	}
	return
}

// DisplayName returns the value of the 'display_name' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *Subscription) DisplayName() string {
	if o != nil && o.displayName != nil {
		return *o.displayName
	}
	return ""
}

// GetDisplayName returns the value of the 'display_name' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *Subscription) GetDisplayName() (value string, ok bool) {
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
func (o *Subscription) ExternalClusterID() string {
	if o != nil && o.externalClusterID != nil {
		return *o.externalClusterID
	}
	return ""
}

// GetExternalClusterID returns the value of the 'external_cluster_ID' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *Subscription) GetExternalClusterID() (value string, ok bool) {
	ok = o != nil && o.externalClusterID != nil
	if ok {
		value = *o.externalClusterID
	}
	return
}

// LastTelemetryDate returns the value of the 'last_telemetry_date' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Last telemetry authorization request for this subscription.
func (o *Subscription) LastTelemetryDate() time.Time {
	if o != nil && o.lastTelemetryDate != nil {
		return *o.lastTelemetryDate
	}
	return time.Time{}
}

// GetLastTelemetryDate returns the value of the 'last_telemetry_date' attribute and
// a flag indicating if the attribute has a value.
//
// Last telemetry authorization request for this subscription.
func (o *Subscription) GetLastTelemetryDate() (value time.Time, ok bool) {
	ok = o != nil && o.lastTelemetryDate != nil
	if ok {
		value = *o.lastTelemetryDate
	}
	return
}

// OrganizationID returns the value of the 'organization_ID' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *Subscription) OrganizationID() string {
	if o != nil && o.organizationID != nil {
		return *o.organizationID
	}
	return ""
}

// GetOrganizationID returns the value of the 'organization_ID' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *Subscription) GetOrganizationID() (value string, ok bool) {
	ok = o != nil && o.organizationID != nil
	if ok {
		value = *o.organizationID
	}
	return
}

// Plan returns the value of the 'plan' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *Subscription) Plan() *Plan {
	if o == nil {
		return nil
	}
	return o.plan
}

// GetPlan returns the value of the 'plan' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *Subscription) GetPlan() (value *Plan, ok bool) {
	ok = o != nil && o.plan != nil
	if ok {
		value = o.plan
	}
	return
}

// RegistryCredential returns the value of the 'registry_credential' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *Subscription) RegistryCredential() *RegistryCredential {
	if o == nil {
		return nil
	}
	return o.registryCredential
}

// GetRegistryCredential returns the value of the 'registry_credential' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *Subscription) GetRegistryCredential() (value *RegistryCredential, ok bool) {
	ok = o != nil && o.registryCredential != nil
	if ok {
		value = o.registryCredential
	}
	return
}

// UpdatedAt returns the value of the 'updated_at' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
//
func (o *Subscription) UpdatedAt() time.Time {
	if o != nil && o.updatedAt != nil {
		return *o.updatedAt
	}
	return time.Time{}
}

// GetUpdatedAt returns the value of the 'updated_at' attribute and
// a flag indicating if the attribute has a value.
//
//
func (o *Subscription) GetUpdatedAt() (value time.Time, ok bool) {
	ok = o != nil && o.updatedAt != nil
	if ok {
		value = *o.updatedAt
	}
	return
}

// SubscriptionListKind is the name of the type used to represent list of
// objects of type 'subscription'.
const SubscriptionListKind = "SubscriptionList"

// SubscriptionListLinkKind is the name of the type used to represent links
// to list of objects of type 'subscription'.
const SubscriptionListLinkKind = "SubscriptionListLink"

// SubscriptionNilKind is the name of the type used to nil lists of
// objects of type 'subscription'.
const SubscriptionListNilKind = "SubscriptionListNil"

// SubscriptionList is a list of values of the 'subscription' type.
type SubscriptionList struct {
	href  *string
	link  bool
	items []*Subscription
}

// Kind returns the name of the type of the object.
func (l *SubscriptionList) Kind() string {
	if l == nil {
		return SubscriptionListNilKind
	}
	if l.link {
		return SubscriptionListLinkKind
	}
	return SubscriptionListKind
}

// Link returns true iif this is a link.
func (l *SubscriptionList) Link() bool {
	return l != nil && l.link
}

// HREF returns the link to the list.
func (l *SubscriptionList) HREF() string {
	if l != nil && l.href != nil {
		return *l.href
	}
	return ""
}

// GetHREF returns the link of the list and a flag indicating if the
// link has a value.
func (l *SubscriptionList) GetHREF() (value string, ok bool) {
	ok = l != nil && l.href != nil
	if ok {
		value = *l.href
	}
	return
}

// Len returns the length of the list.
func (l *SubscriptionList) Len() int {
	if l == nil {
		return 0
	}
	return len(l.items)
}

// Empty returns true if the list is empty.
func (l *SubscriptionList) Empty() bool {
	return l == nil || len(l.items) == 0
}

// Get returns the item of the list with the given index. If there is no item with
// that index it returns nil.
func (l *SubscriptionList) Get(i int) *Subscription {
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
func (l *SubscriptionList) Slice() []*Subscription {
	var slice []*Subscription
	if l == nil {
		slice = make([]*Subscription, 0)
	} else {
		slice = make([]*Subscription, len(l.items))
		copy(slice, l.items)
	}
	return slice
}

// Each runs the given function for each item of the list, in order. If the function
// returns false the iteration stops, otherwise it continues till all the elements
// of the list have been processed.
func (l *SubscriptionList) Each(f func(item *Subscription) bool) {
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
func (l *SubscriptionList) Range(f func(index int, item *Subscription) bool) {
	if l == nil {
		return
	}
	for index, item := range l.items {
		if !f(index, item) {
			break
		}
	}
}
