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

// RoleBindingBuilder contains the data and logic needed to build 'role_binding' objects.
//
//
type RoleBindingBuilder struct {
	id            *string
	href          *string
	link          bool
	account       *AccountBuilder
	configManaged *bool
	organization  *OrganizationBuilder
	role          *RoleBuilder
	subscription  *SubscriptionBuilder
	type_         *string
}

// NewRoleBinding creates a new builder of 'role_binding' objects.
func NewRoleBinding() *RoleBindingBuilder {
	return new(RoleBindingBuilder)
}

// ID sets the identifier of the object.
func (b *RoleBindingBuilder) ID(value string) *RoleBindingBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *RoleBindingBuilder) HREF(value string) *RoleBindingBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *RoleBindingBuilder) Link(value bool) *RoleBindingBuilder {
	b.link = value
	return b
}

// Account sets the value of the 'account' attribute
// to the given value.
//
//
func (b *RoleBindingBuilder) Account(value *AccountBuilder) *RoleBindingBuilder {
	b.account = value
	return b
}

// ConfigManaged sets the value of the 'config_managed' attribute
// to the given value.
//
//
func (b *RoleBindingBuilder) ConfigManaged(value bool) *RoleBindingBuilder {
	b.configManaged = &value
	return b
}

// Organization sets the value of the 'organization' attribute
// to the given value.
//
//
func (b *RoleBindingBuilder) Organization(value *OrganizationBuilder) *RoleBindingBuilder {
	b.organization = value
	return b
}

// Role sets the value of the 'role' attribute
// to the given value.
//
//
func (b *RoleBindingBuilder) Role(value *RoleBuilder) *RoleBindingBuilder {
	b.role = value
	return b
}

// Subscription sets the value of the 'subscription' attribute
// to the given value.
//
//
func (b *RoleBindingBuilder) Subscription(value *SubscriptionBuilder) *RoleBindingBuilder {
	b.subscription = value
	return b
}

// Type sets the value of the 'type' attribute
// to the given value.
//
//
func (b *RoleBindingBuilder) Type(value string) *RoleBindingBuilder {
	b.type_ = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *RoleBindingBuilder) Copy(object *RoleBinding) *RoleBindingBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	if object.account != nil {
		b.account = NewAccount().Copy(object.account)
	} else {
		b.account = nil
	}
	b.configManaged = object.configManaged
	if object.organization != nil {
		b.organization = NewOrganization().Copy(object.organization)
	} else {
		b.organization = nil
	}
	if object.role != nil {
		b.role = NewRole().Copy(object.role)
	} else {
		b.role = nil
	}
	if object.subscription != nil {
		b.subscription = NewSubscription().Copy(object.subscription)
	} else {
		b.subscription = nil
	}
	b.type_ = object.type_
	return b
}

// Build creates a 'role_binding' object using the configuration stored in the builder.
func (b *RoleBindingBuilder) Build() (object *RoleBinding, err error) {
	object = new(RoleBinding)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	if b.account != nil {
		object.account, err = b.account.Build()
		if err != nil {
			return
		}
	}
	if b.configManaged != nil {
		object.configManaged = b.configManaged
	}
	if b.organization != nil {
		object.organization, err = b.organization.Build()
		if err != nil {
			return
		}
	}
	if b.role != nil {
		object.role, err = b.role.Build()
		if err != nil {
			return
		}
	}
	if b.subscription != nil {
		object.subscription, err = b.subscription.Build()
		if err != nil {
			return
		}
	}
	if b.type_ != nil {
		object.type_ = b.type_
	}
	return
}
