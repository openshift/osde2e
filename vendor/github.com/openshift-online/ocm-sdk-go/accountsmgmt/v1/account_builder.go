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

// AccountBuilder contains the data and logic needed to build 'account' objects.
//
//
type AccountBuilder struct {
	id             *string
	href           *string
	link           bool
	banCode        *string
	banDescription *string
	banned         *bool
	createdAt      *time.Time
	email          *string
	firstName      *string
	lastName       *string
	organization   *OrganizationBuilder
	serviceAccount *bool
	updatedAt      *time.Time
	username       *string
}

// NewAccount creates a new builder of 'account' objects.
func NewAccount() *AccountBuilder {
	return new(AccountBuilder)
}

// ID sets the identifier of the object.
func (b *AccountBuilder) ID(value string) *AccountBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *AccountBuilder) HREF(value string) *AccountBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *AccountBuilder) Link(value bool) *AccountBuilder {
	b.link = value
	return b
}

// BanCode sets the value of the 'ban_code' attribute to the given value.
//
//
func (b *AccountBuilder) BanCode(value string) *AccountBuilder {
	b.banCode = &value
	return b
}

// BanDescription sets the value of the 'ban_description' attribute to the given value.
//
//
func (b *AccountBuilder) BanDescription(value string) *AccountBuilder {
	b.banDescription = &value
	return b
}

// Banned sets the value of the 'banned' attribute to the given value.
//
//
func (b *AccountBuilder) Banned(value bool) *AccountBuilder {
	b.banned = &value
	return b
}

// CreatedAt sets the value of the 'created_at' attribute to the given value.
//
//
func (b *AccountBuilder) CreatedAt(value time.Time) *AccountBuilder {
	b.createdAt = &value
	return b
}

// Email sets the value of the 'email' attribute to the given value.
//
//
func (b *AccountBuilder) Email(value string) *AccountBuilder {
	b.email = &value
	return b
}

// FirstName sets the value of the 'first_name' attribute to the given value.
//
//
func (b *AccountBuilder) FirstName(value string) *AccountBuilder {
	b.firstName = &value
	return b
}

// LastName sets the value of the 'last_name' attribute to the given value.
//
//
func (b *AccountBuilder) LastName(value string) *AccountBuilder {
	b.lastName = &value
	return b
}

// Organization sets the value of the 'organization' attribute to the given value.
//
//
func (b *AccountBuilder) Organization(value *OrganizationBuilder) *AccountBuilder {
	b.organization = value
	return b
}

// ServiceAccount sets the value of the 'service_account' attribute to the given value.
//
//
func (b *AccountBuilder) ServiceAccount(value bool) *AccountBuilder {
	b.serviceAccount = &value
	return b
}

// UpdatedAt sets the value of the 'updated_at' attribute to the given value.
//
//
func (b *AccountBuilder) UpdatedAt(value time.Time) *AccountBuilder {
	b.updatedAt = &value
	return b
}

// Username sets the value of the 'username' attribute to the given value.
//
//
func (b *AccountBuilder) Username(value string) *AccountBuilder {
	b.username = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *AccountBuilder) Copy(object *Account) *AccountBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.banCode = object.banCode
	b.banDescription = object.banDescription
	b.banned = object.banned
	b.createdAt = object.createdAt
	b.email = object.email
	b.firstName = object.firstName
	b.lastName = object.lastName
	if object.organization != nil {
		b.organization = NewOrganization().Copy(object.organization)
	} else {
		b.organization = nil
	}
	b.serviceAccount = object.serviceAccount
	b.updatedAt = object.updatedAt
	b.username = object.username
	return b
}

// Build creates a 'account' object using the configuration stored in the builder.
func (b *AccountBuilder) Build() (object *Account, err error) {
	object = new(Account)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	object.banCode = b.banCode
	object.banDescription = b.banDescription
	object.banned = b.banned
	object.createdAt = b.createdAt
	object.email = b.email
	object.firstName = b.firstName
	object.lastName = b.lastName
	if b.organization != nil {
		object.organization, err = b.organization.Build()
		if err != nil {
			return
		}
	}
	object.serviceAccount = b.serviceAccount
	object.updatedAt = b.updatedAt
	object.username = b.username
	return
}
