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

// VersionBuilder contains the data and logic needed to build 'version' objects.
//
// Representation of an _OpenShift_ version.
type VersionBuilder struct {
	id       *string
	href     *string
	link     bool
	default_ *bool
	enabled  *bool
}

// NewVersion creates a new builder of 'version' objects.
func NewVersion() *VersionBuilder {
	return new(VersionBuilder)
}

// ID sets the identifier of the object.
func (b *VersionBuilder) ID(value string) *VersionBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *VersionBuilder) HREF(value string) *VersionBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *VersionBuilder) Link(value bool) *VersionBuilder {
	b.link = value
	return b
}

// Default sets the value of the 'default' attribute to the given value.
//
//
func (b *VersionBuilder) Default(value bool) *VersionBuilder {
	b.default_ = &value
	return b
}

// Enabled sets the value of the 'enabled' attribute to the given value.
//
//
func (b *VersionBuilder) Enabled(value bool) *VersionBuilder {
	b.enabled = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *VersionBuilder) Copy(object *Version) *VersionBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.default_ = object.default_
	b.enabled = object.enabled
	return b
}

// Build creates a 'version' object using the configuration stored in the builder.
func (b *VersionBuilder) Build() (object *Version, err error) {
	object = new(Version)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	object.default_ = b.default_
	object.enabled = b.enabled
	return
}
