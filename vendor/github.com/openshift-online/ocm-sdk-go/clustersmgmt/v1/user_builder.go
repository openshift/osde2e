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

// UserBuilder contains the data and logic needed to build 'user' objects.
//
// Representation of a user.
type UserBuilder struct {
	id   *string
	href *string
	link bool
}

// NewUser creates a new builder of 'user' objects.
func NewUser() *UserBuilder {
	return new(UserBuilder)
}

// ID sets the identifier of the object.
func (b *UserBuilder) ID(value string) *UserBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *UserBuilder) HREF(value string) *UserBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *UserBuilder) Link(value bool) *UserBuilder {
	b.link = value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *UserBuilder) Copy(object *User) *UserBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	return b
}

// Build creates a 'user' object using the configuration stored in the builder.
func (b *UserBuilder) Build() (object *User, err error) {
	object = new(User)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	return
}
