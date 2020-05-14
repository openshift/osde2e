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

// PlanBuilder contains the data and logic needed to build 'plan' objects.
//
//
type PlanBuilder struct {
	id    *string
	href  *string
	link  bool
	name  *string
	type_ *string
}

// NewPlan creates a new builder of 'plan' objects.
func NewPlan() *PlanBuilder {
	return new(PlanBuilder)
}

// ID sets the identifier of the object.
func (b *PlanBuilder) ID(value string) *PlanBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *PlanBuilder) HREF(value string) *PlanBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *PlanBuilder) Link(value bool) *PlanBuilder {
	b.link = value
	return b
}

// Name sets the value of the 'name' attribute to the given value.
//
//
func (b *PlanBuilder) Name(value string) *PlanBuilder {
	b.name = &value
	return b
}

// Type sets the value of the 'type' attribute to the given value.
//
//
func (b *PlanBuilder) Type(value string) *PlanBuilder {
	b.type_ = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *PlanBuilder) Copy(object *Plan) *PlanBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.name = object.name
	b.type_ = object.type_
	return b
}

// Build creates a 'plan' object using the configuration stored in the builder.
func (b *PlanBuilder) Build() (object *Plan, err error) {
	object = new(Plan)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	object.name = b.name
	object.type_ = b.type_
	return
}
