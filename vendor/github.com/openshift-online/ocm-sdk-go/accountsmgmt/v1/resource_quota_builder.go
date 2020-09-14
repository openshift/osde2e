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

// ResourceQuotaBuilder contains the data and logic needed to build 'resource_quota' objects.
//
//
type ResourceQuotaBuilder struct {
	id                   *string
	href                 *string
	link                 bool
	byoc                 *bool
	sku                  *string
	allowed              *int
	availabilityZoneType *string
	createdAt            *time.Time
	organizationID       *string
	resourceName         *string
	resourceType         *string
	skuCount             *int
	type_                *string
	updatedAt            *time.Time
}

// NewResourceQuota creates a new builder of 'resource_quota' objects.
func NewResourceQuota() *ResourceQuotaBuilder {
	return new(ResourceQuotaBuilder)
}

// ID sets the identifier of the object.
func (b *ResourceQuotaBuilder) ID(value string) *ResourceQuotaBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *ResourceQuotaBuilder) HREF(value string) *ResourceQuotaBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *ResourceQuotaBuilder) Link(value bool) *ResourceQuotaBuilder {
	b.link = value
	return b
}

// BYOC sets the value of the 'BYOC' attribute to the given value.
//
//
func (b *ResourceQuotaBuilder) BYOC(value bool) *ResourceQuotaBuilder {
	b.byoc = &value
	return b
}

// SKU sets the value of the 'SKU' attribute to the given value.
//
//
func (b *ResourceQuotaBuilder) SKU(value string) *ResourceQuotaBuilder {
	b.sku = &value
	return b
}

// Allowed sets the value of the 'allowed' attribute to the given value.
//
//
func (b *ResourceQuotaBuilder) Allowed(value int) *ResourceQuotaBuilder {
	b.allowed = &value
	return b
}

// AvailabilityZoneType sets the value of the 'availability_zone_type' attribute to the given value.
//
//
func (b *ResourceQuotaBuilder) AvailabilityZoneType(value string) *ResourceQuotaBuilder {
	b.availabilityZoneType = &value
	return b
}

// CreatedAt sets the value of the 'created_at' attribute to the given value.
//
//
func (b *ResourceQuotaBuilder) CreatedAt(value time.Time) *ResourceQuotaBuilder {
	b.createdAt = &value
	return b
}

// OrganizationID sets the value of the 'organization_ID' attribute to the given value.
//
//
func (b *ResourceQuotaBuilder) OrganizationID(value string) *ResourceQuotaBuilder {
	b.organizationID = &value
	return b
}

// ResourceName sets the value of the 'resource_name' attribute to the given value.
//
//
func (b *ResourceQuotaBuilder) ResourceName(value string) *ResourceQuotaBuilder {
	b.resourceName = &value
	return b
}

// ResourceType sets the value of the 'resource_type' attribute to the given value.
//
//
func (b *ResourceQuotaBuilder) ResourceType(value string) *ResourceQuotaBuilder {
	b.resourceType = &value
	return b
}

// SkuCount sets the value of the 'sku_count' attribute to the given value.
//
//
func (b *ResourceQuotaBuilder) SkuCount(value int) *ResourceQuotaBuilder {
	b.skuCount = &value
	return b
}

// Type sets the value of the 'type' attribute to the given value.
//
//
func (b *ResourceQuotaBuilder) Type(value string) *ResourceQuotaBuilder {
	b.type_ = &value
	return b
}

// UpdatedAt sets the value of the 'updated_at' attribute to the given value.
//
//
func (b *ResourceQuotaBuilder) UpdatedAt(value time.Time) *ResourceQuotaBuilder {
	b.updatedAt = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *ResourceQuotaBuilder) Copy(object *ResourceQuota) *ResourceQuotaBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.byoc = object.byoc
	b.sku = object.sku
	b.allowed = object.allowed
	b.availabilityZoneType = object.availabilityZoneType
	b.createdAt = object.createdAt
	b.organizationID = object.organizationID
	b.resourceName = object.resourceName
	b.resourceType = object.resourceType
	b.skuCount = object.skuCount
	b.type_ = object.type_
	b.updatedAt = object.updatedAt
	return b
}

// Build creates a 'resource_quota' object using the configuration stored in the builder.
func (b *ResourceQuotaBuilder) Build() (object *ResourceQuota, err error) {
	object = new(ResourceQuota)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	object.byoc = b.byoc
	object.sku = b.sku
	object.allowed = b.allowed
	object.availabilityZoneType = b.availabilityZoneType
	object.createdAt = b.createdAt
	object.organizationID = b.organizationID
	object.resourceName = b.resourceName
	object.resourceType = b.resourceType
	object.skuCount = b.skuCount
	object.type_ = b.type_
	object.updatedAt = b.updatedAt
	return
}
