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

// OrganizationBuilder contains the data and logic needed to build 'organization' objects.
//
//
type OrganizationBuilder struct {
	id           *string
	href         *string
	link         bool
	createdAt    *time.Time
	ebsAccountID *string
	externalID   *string
	labels       []*LabelBuilder
	name         *string
	updatedAt    *time.Time
}

// NewOrganization creates a new builder of 'organization' objects.
func NewOrganization() *OrganizationBuilder {
	return new(OrganizationBuilder)
}

// ID sets the identifier of the object.
func (b *OrganizationBuilder) ID(value string) *OrganizationBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *OrganizationBuilder) HREF(value string) *OrganizationBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *OrganizationBuilder) Link(value bool) *OrganizationBuilder {
	b.link = value
	return b
}

// CreatedAt sets the value of the 'created_at' attribute to the given value.
//
//
func (b *OrganizationBuilder) CreatedAt(value time.Time) *OrganizationBuilder {
	b.createdAt = &value
	return b
}

// EbsAccountID sets the value of the 'ebs_account_ID' attribute to the given value.
//
//
func (b *OrganizationBuilder) EbsAccountID(value string) *OrganizationBuilder {
	b.ebsAccountID = &value
	return b
}

// ExternalID sets the value of the 'external_ID' attribute to the given value.
//
//
func (b *OrganizationBuilder) ExternalID(value string) *OrganizationBuilder {
	b.externalID = &value
	return b
}

// Labels sets the value of the 'labels' attribute to the given values.
//
//
func (b *OrganizationBuilder) Labels(values ...*LabelBuilder) *OrganizationBuilder {
	b.labels = make([]*LabelBuilder, len(values))
	copy(b.labels, values)
	return b
}

// Name sets the value of the 'name' attribute to the given value.
//
//
func (b *OrganizationBuilder) Name(value string) *OrganizationBuilder {
	b.name = &value
	return b
}

// UpdatedAt sets the value of the 'updated_at' attribute to the given value.
//
//
func (b *OrganizationBuilder) UpdatedAt(value time.Time) *OrganizationBuilder {
	b.updatedAt = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *OrganizationBuilder) Copy(object *Organization) *OrganizationBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.createdAt = object.createdAt
	b.ebsAccountID = object.ebsAccountID
	b.externalID = object.externalID
	if object.labels != nil {
		b.labels = make([]*LabelBuilder, len(object.labels))
		for i, v := range object.labels {
			b.labels[i] = NewLabel().Copy(v)
		}
	} else {
		b.labels = nil
	}
	b.name = object.name
	b.updatedAt = object.updatedAt
	return b
}

// Build creates a 'organization' object using the configuration stored in the builder.
func (b *OrganizationBuilder) Build() (object *Organization, err error) {
	object = new(Organization)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	object.createdAt = b.createdAt
	object.ebsAccountID = b.ebsAccountID
	object.externalID = b.externalID
	if b.labels != nil {
		object.labels = make([]*Label, len(b.labels))
		for i, v := range b.labels {
			object.labels[i], err = v.Build()
			if err != nil {
				return
			}
		}
	}
	object.name = b.name
	object.updatedAt = b.updatedAt
	return
}
