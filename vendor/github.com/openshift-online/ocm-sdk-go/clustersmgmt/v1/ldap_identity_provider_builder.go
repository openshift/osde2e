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

package v1 // github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1

// LDAPIdentityProviderBuilder contains the data and logic needed to build 'LDAP_identity_provider' objects.
//
// Details for `ldap` identity providers.
type LDAPIdentityProviderBuilder struct {
	ca             *string
	ldapAttributes *LDAPAttributesBuilder
	url            *string
	bindDN         *string
	bindPassword   *string
	insecure       *bool
}

// NewLDAPIdentityProvider creates a new builder of 'LDAP_identity_provider' objects.
func NewLDAPIdentityProvider() *LDAPIdentityProviderBuilder {
	return new(LDAPIdentityProviderBuilder)
}

// CA sets the value of the 'CA' attribute
// to the given value.
//
//
func (b *LDAPIdentityProviderBuilder) CA(value string) *LDAPIdentityProviderBuilder {
	b.ca = &value
	return b
}

// LDAPAttributes sets the value of the 'LDAP_attributes' attribute
// to the given value.
//
// LDAP attributes used to configure the LDAP identity provider.
func (b *LDAPIdentityProviderBuilder) LDAPAttributes(value *LDAPAttributesBuilder) *LDAPIdentityProviderBuilder {
	b.ldapAttributes = value
	return b
}

// URL sets the value of the 'URL' attribute
// to the given value.
//
//
func (b *LDAPIdentityProviderBuilder) URL(value string) *LDAPIdentityProviderBuilder {
	b.url = &value
	return b
}

// BindDN sets the value of the 'bind_DN' attribute
// to the given value.
//
//
func (b *LDAPIdentityProviderBuilder) BindDN(value string) *LDAPIdentityProviderBuilder {
	b.bindDN = &value
	return b
}

// BindPassword sets the value of the 'bind_password' attribute
// to the given value.
//
//
func (b *LDAPIdentityProviderBuilder) BindPassword(value string) *LDAPIdentityProviderBuilder {
	b.bindPassword = &value
	return b
}

// Insecure sets the value of the 'insecure' attribute
// to the given value.
//
//
func (b *LDAPIdentityProviderBuilder) Insecure(value bool) *LDAPIdentityProviderBuilder {
	b.insecure = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *LDAPIdentityProviderBuilder) Copy(object *LDAPIdentityProvider) *LDAPIdentityProviderBuilder {
	if object == nil {
		return b
	}
	b.ca = object.ca
	if object.ldapAttributes != nil {
		b.ldapAttributes = NewLDAPAttributes().Copy(object.ldapAttributes)
	} else {
		b.ldapAttributes = nil
	}
	b.url = object.url
	b.bindDN = object.bindDN
	b.bindPassword = object.bindPassword
	b.insecure = object.insecure
	return b
}

// Build creates a 'LDAP_identity_provider' object using the configuration stored in the builder.
func (b *LDAPIdentityProviderBuilder) Build() (object *LDAPIdentityProvider, err error) {
	object = new(LDAPIdentityProvider)
	if b.ca != nil {
		object.ca = b.ca
	}
	if b.ldapAttributes != nil {
		object.ldapAttributes, err = b.ldapAttributes.Build()
		if err != nil {
			return
		}
	}
	if b.url != nil {
		object.url = b.url
	}
	if b.bindDN != nil {
		object.bindDN = b.bindDN
	}
	if b.bindPassword != nil {
		object.bindPassword = b.bindPassword
	}
	if b.insecure != nil {
		object.insecure = b.insecure
	}
	return
}
