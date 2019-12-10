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

// QuotaSummaryBuilder contains the data and logic needed to build 'quota_summary' objects.
//
//
type QuotaSummaryBuilder struct {
	byoc                 *bool
	allowed              *int
	availabilityZoneType *string
	organizationID       *string
	reserved             *int
	resourceName         *string
	resourceType         *string
}

// NewQuotaSummary creates a new builder of 'quota_summary' objects.
func NewQuotaSummary() *QuotaSummaryBuilder {
	return new(QuotaSummaryBuilder)
}

// BYOC sets the value of the 'BYOC' attribute
// to the given value.
//
//
func (b *QuotaSummaryBuilder) BYOC(value bool) *QuotaSummaryBuilder {
	b.byoc = &value
	return b
}

// Allowed sets the value of the 'allowed' attribute
// to the given value.
//
//
func (b *QuotaSummaryBuilder) Allowed(value int) *QuotaSummaryBuilder {
	b.allowed = &value
	return b
}

// AvailabilityZoneType sets the value of the 'availability_zone_type' attribute
// to the given value.
//
//
func (b *QuotaSummaryBuilder) AvailabilityZoneType(value string) *QuotaSummaryBuilder {
	b.availabilityZoneType = &value
	return b
}

// OrganizationID sets the value of the 'organization_ID' attribute
// to the given value.
//
//
func (b *QuotaSummaryBuilder) OrganizationID(value string) *QuotaSummaryBuilder {
	b.organizationID = &value
	return b
}

// Reserved sets the value of the 'reserved' attribute
// to the given value.
//
//
func (b *QuotaSummaryBuilder) Reserved(value int) *QuotaSummaryBuilder {
	b.reserved = &value
	return b
}

// ResourceName sets the value of the 'resource_name' attribute
// to the given value.
//
//
func (b *QuotaSummaryBuilder) ResourceName(value string) *QuotaSummaryBuilder {
	b.resourceName = &value
	return b
}

// ResourceType sets the value of the 'resource_type' attribute
// to the given value.
//
//
func (b *QuotaSummaryBuilder) ResourceType(value string) *QuotaSummaryBuilder {
	b.resourceType = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *QuotaSummaryBuilder) Copy(object *QuotaSummary) *QuotaSummaryBuilder {
	if object == nil {
		return b
	}
	b.byoc = object.byoc
	b.allowed = object.allowed
	b.availabilityZoneType = object.availabilityZoneType
	b.organizationID = object.organizationID
	b.reserved = object.reserved
	b.resourceName = object.resourceName
	b.resourceType = object.resourceType
	return b
}

// Build creates a 'quota_summary' object using the configuration stored in the builder.
func (b *QuotaSummaryBuilder) Build() (object *QuotaSummary, err error) {
	object = new(QuotaSummary)
	if b.byoc != nil {
		object.byoc = b.byoc
	}
	if b.allowed != nil {
		object.allowed = b.allowed
	}
	if b.availabilityZoneType != nil {
		object.availabilityZoneType = b.availabilityZoneType
	}
	if b.organizationID != nil {
		object.organizationID = b.organizationID
	}
	if b.reserved != nil {
		object.reserved = b.reserved
	}
	if b.resourceName != nil {
		object.resourceName = b.resourceName
	}
	if b.resourceType != nil {
		object.resourceType = b.resourceType
	}
	return
}
