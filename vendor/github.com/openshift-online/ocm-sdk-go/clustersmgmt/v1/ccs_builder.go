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

// CCSBuilder contains the data and logic needed to build 'CCS' objects.
//
//
type CCSBuilder struct {
	id               *string
	href             *string
	link             bool
	disableSCPChecks *bool
	enabled          *bool
}

// NewCCS creates a new builder of 'CCS' objects.
func NewCCS() *CCSBuilder {
	return new(CCSBuilder)
}

// ID sets the identifier of the object.
func (b *CCSBuilder) ID(value string) *CCSBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *CCSBuilder) HREF(value string) *CCSBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *CCSBuilder) Link(value bool) *CCSBuilder {
	b.link = value
	return b
}

// DisableSCPChecks sets the value of the 'disable_SCP_checks' attribute to the given value.
//
//
func (b *CCSBuilder) DisableSCPChecks(value bool) *CCSBuilder {
	b.disableSCPChecks = &value
	return b
}

// Enabled sets the value of the 'enabled' attribute to the given value.
//
//
func (b *CCSBuilder) Enabled(value bool) *CCSBuilder {
	b.enabled = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *CCSBuilder) Copy(object *CCS) *CCSBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.disableSCPChecks = object.disableSCPChecks
	b.enabled = object.enabled
	return b
}

// Build creates a 'CCS' object using the configuration stored in the builder.
func (b *CCSBuilder) Build() (object *CCS, err error) {
	object = new(CCS)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	object.disableSCPChecks = b.disableSCPChecks
	object.enabled = b.enabled
	return
}
