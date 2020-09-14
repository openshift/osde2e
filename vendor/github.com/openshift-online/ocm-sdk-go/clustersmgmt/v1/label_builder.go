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

// LabelBuilder contains the data and logic needed to build 'label' objects.
//
// Representation of a label in clusterdeployment.
type LabelBuilder struct {
	id    *string
	href  *string
	link  bool
	key   *string
	value *string
}

// NewLabel creates a new builder of 'label' objects.
func NewLabel() *LabelBuilder {
	return new(LabelBuilder)
}

// ID sets the identifier of the object.
func (b *LabelBuilder) ID(value string) *LabelBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *LabelBuilder) HREF(value string) *LabelBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *LabelBuilder) Link(value bool) *LabelBuilder {
	b.link = value
	return b
}

// Key sets the value of the 'key' attribute to the given value.
//
//
func (b *LabelBuilder) Key(value string) *LabelBuilder {
	b.key = &value
	return b
}

// Value sets the value of the 'value' attribute to the given value.
//
//
func (b *LabelBuilder) Value(value string) *LabelBuilder {
	b.value = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *LabelBuilder) Copy(object *Label) *LabelBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.key = object.key
	b.value = object.value
	return b
}

// Build creates a 'label' object using the configuration stored in the builder.
func (b *LabelBuilder) Build() (object *Label, err error) {
	object = new(Label)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	object.key = b.key
	object.value = b.value
	return
}
