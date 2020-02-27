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

// SKUBuilder contains the data and logic needed to build 'SKU' objects.
//
// Identifies computing resources
type SKUBuilder struct {
	id                   *string
	href                 *string
	link                 bool
	byoc                 *bool
	availabilityZoneType *string
	resourceName         *string
	resourceType         *string
	resources            []*ResourceBuilder
}

// NewSKU creates a new builder of 'SKU' objects.
func NewSKU() *SKUBuilder {
	return new(SKUBuilder)
}

// ID sets the identifier of the object.
func (b *SKUBuilder) ID(value string) *SKUBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *SKUBuilder) HREF(value string) *SKUBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *SKUBuilder) Link(value bool) *SKUBuilder {
	b.link = value
	return b
}

// BYOC sets the value of the 'BYOC' attribute
// to the given value.
//
//
func (b *SKUBuilder) BYOC(value bool) *SKUBuilder {
	b.byoc = &value
	return b
}

// AvailabilityZoneType sets the value of the 'availability_zone_type' attribute
// to the given value.
//
//
func (b *SKUBuilder) AvailabilityZoneType(value string) *SKUBuilder {
	b.availabilityZoneType = &value
	return b
}

// ResourceName sets the value of the 'resource_name' attribute
// to the given value.
//
//
func (b *SKUBuilder) ResourceName(value string) *SKUBuilder {
	b.resourceName = &value
	return b
}

// ResourceType sets the value of the 'resource_type' attribute
// to the given value.
//
//
func (b *SKUBuilder) ResourceType(value string) *SKUBuilder {
	b.resourceType = &value
	return b
}

// Resources sets the value of the 'resources' attribute
// to the given values.
//
//
func (b *SKUBuilder) Resources(values ...*ResourceBuilder) *SKUBuilder {
	b.resources = make([]*ResourceBuilder, len(values))
	copy(b.resources, values)
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *SKUBuilder) Copy(object *SKU) *SKUBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.byoc = object.byoc
	b.availabilityZoneType = object.availabilityZoneType
	b.resourceName = object.resourceName
	b.resourceType = object.resourceType
	if object.resources != nil && len(object.resources.items) > 0 {
		b.resources = make([]*ResourceBuilder, len(object.resources.items))
		for i, item := range object.resources.items {
			b.resources[i] = NewResource().Copy(item)
		}
	} else {
		b.resources = nil
	}
	return b
}

// Build creates a 'SKU' object using the configuration stored in the builder.
func (b *SKUBuilder) Build() (object *SKU, err error) {
	object = new(SKU)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	if b.byoc != nil {
		object.byoc = b.byoc
	}
	if b.availabilityZoneType != nil {
		object.availabilityZoneType = b.availabilityZoneType
	}
	if b.resourceName != nil {
		object.resourceName = b.resourceName
	}
	if b.resourceType != nil {
		object.resourceType = b.resourceType
	}
	if b.resources != nil {
		object.resources = new(ResourceList)
		object.resources.items = make([]*Resource, len(b.resources))
		for i, item := range b.resources {
			object.resources.items[i], err = item.Build()
			if err != nil {
				return
			}
		}
	}
	return
}
