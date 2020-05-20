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

import (
	time "time"
)

// ReservedResourceBuilder contains the data and logic needed to build 'reserved_resource' objects.
//
//
type ReservedResourceBuilder struct {
	byoc                 *bool
	availabilityZoneType *string
	count                *int
	createdAt            *time.Time
	resourceName         *string
	resourceType         *string
	updatedAt            *time.Time
}

// NewReservedResource creates a new builder of 'reserved_resource' objects.
func NewReservedResource() *ReservedResourceBuilder {
	return new(ReservedResourceBuilder)
}

// BYOC sets the value of the 'BYOC' attribute to the given value.
//
//
func (b *ReservedResourceBuilder) BYOC(value bool) *ReservedResourceBuilder {
	b.byoc = &value
	return b
}

// AvailabilityZoneType sets the value of the 'availability_zone_type' attribute to the given value.
//
//
func (b *ReservedResourceBuilder) AvailabilityZoneType(value string) *ReservedResourceBuilder {
	b.availabilityZoneType = &value
	return b
}

// Count sets the value of the 'count' attribute to the given value.
//
//
func (b *ReservedResourceBuilder) Count(value int) *ReservedResourceBuilder {
	b.count = &value
	return b
}

// CreatedAt sets the value of the 'created_at' attribute to the given value.
//
//
func (b *ReservedResourceBuilder) CreatedAt(value time.Time) *ReservedResourceBuilder {
	b.createdAt = &value
	return b
}

// ResourceName sets the value of the 'resource_name' attribute to the given value.
//
//
func (b *ReservedResourceBuilder) ResourceName(value string) *ReservedResourceBuilder {
	b.resourceName = &value
	return b
}

// ResourceType sets the value of the 'resource_type' attribute to the given value.
//
//
func (b *ReservedResourceBuilder) ResourceType(value string) *ReservedResourceBuilder {
	b.resourceType = &value
	return b
}

// UpdatedAt sets the value of the 'updated_at' attribute to the given value.
//
//
func (b *ReservedResourceBuilder) UpdatedAt(value time.Time) *ReservedResourceBuilder {
	b.updatedAt = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *ReservedResourceBuilder) Copy(object *ReservedResource) *ReservedResourceBuilder {
	if object == nil {
		return b
	}
	b.byoc = object.byoc
	b.availabilityZoneType = object.availabilityZoneType
	b.count = object.count
	b.createdAt = object.createdAt
	b.resourceName = object.resourceName
	b.resourceType = object.resourceType
	b.updatedAt = object.updatedAt
	return b
}

// Build creates a 'reserved_resource' object using the configuration stored in the builder.
func (b *ReservedResourceBuilder) Build() (object *ReservedResource, err error) {
	object = new(ReservedResource)
	object.byoc = b.byoc
	object.availabilityZoneType = b.availabilityZoneType
	object.count = b.count
	object.createdAt = b.createdAt
	object.resourceName = b.resourceName
	object.resourceType = b.resourceType
	object.updatedAt = b.updatedAt
	return
}
