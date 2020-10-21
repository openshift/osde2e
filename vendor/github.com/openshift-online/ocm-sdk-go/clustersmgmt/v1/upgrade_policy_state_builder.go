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

// UpgradePolicyStateBuilder contains the data and logic needed to build 'upgrade_policy_state' objects.
//
// Representation of an upgrade policy state that that is set for a cluster.
type UpgradePolicyStateBuilder struct {
	id          *string
	href        *string
	link        bool
	description *string
	value       *string
}

// NewUpgradePolicyState creates a new builder of 'upgrade_policy_state' objects.
func NewUpgradePolicyState() *UpgradePolicyStateBuilder {
	return new(UpgradePolicyStateBuilder)
}

// ID sets the identifier of the object.
func (b *UpgradePolicyStateBuilder) ID(value string) *UpgradePolicyStateBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *UpgradePolicyStateBuilder) HREF(value string) *UpgradePolicyStateBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *UpgradePolicyStateBuilder) Link(value bool) *UpgradePolicyStateBuilder {
	b.link = value
	return b
}

// Description sets the value of the 'description' attribute to the given value.
//
//
func (b *UpgradePolicyStateBuilder) Description(value string) *UpgradePolicyStateBuilder {
	b.description = &value
	return b
}

// Value sets the value of the 'value' attribute to the given value.
//
//
func (b *UpgradePolicyStateBuilder) Value(value string) *UpgradePolicyStateBuilder {
	b.value = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *UpgradePolicyStateBuilder) Copy(object *UpgradePolicyState) *UpgradePolicyStateBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.description = object.description
	b.value = object.value
	return b
}

// Build creates a 'upgrade_policy_state' object using the configuration stored in the builder.
func (b *UpgradePolicyStateBuilder) Build() (object *UpgradePolicyState, err error) {
	object = new(UpgradePolicyState)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	object.description = b.description
	object.value = b.value
	return
}
