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

import (
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// ldapIdentityProviderData is the data structure used internally to marshal and unmarshal
// objects of type 'LDAP_identity_provider'.
type ldapIdentityProviderData struct {
	CA             *string             "json:\"ca,omitempty\""
	LDAPAttributes *ldapAttributesData "json:\"ldap_attributes,omitempty\""
	URL            *string             "json:\"url,omitempty\""
	BindDN         *string             "json:\"bind_dn,omitempty\""
	BindPassword   *string             "json:\"bind_password,omitempty\""
	Insecure       *bool               "json:\"insecure,omitempty\""
}

// MarshalLDAPIdentityProvider writes a value of the 'LDAP_identity_provider' to the given target,
// which can be a writer or a JSON encoder.
func MarshalLDAPIdentityProvider(object *LDAPIdentityProvider, target interface{}) error {
	encoder, err := helpers.NewEncoder(target)
	if err != nil {
		return err
	}
	data, err := object.wrap()
	if err != nil {
		return err
	}
	return encoder.Encode(data)
}

// wrap is the method used internally to convert a value of the 'LDAP_identity_provider'
// value to a JSON document.
func (o *LDAPIdentityProvider) wrap() (data *ldapIdentityProviderData, err error) {
	if o == nil {
		return
	}
	data = new(ldapIdentityProviderData)
	data.CA = o.ca
	data.LDAPAttributes, err = o.ldapAttributes.wrap()
	if err != nil {
		return
	}
	data.URL = o.url
	data.BindDN = o.bindDN
	data.BindPassword = o.bindPassword
	data.Insecure = o.insecure
	return
}

// UnmarshalLDAPIdentityProvider reads a value of the 'LDAP_identity_provider' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalLDAPIdentityProvider(source interface{}) (object *LDAPIdentityProvider, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(ldapIdentityProviderData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'LDAP_identity_provider' type.
func (d *ldapIdentityProviderData) unwrap() (object *LDAPIdentityProvider, err error) {
	if d == nil {
		return
	}
	object = new(LDAPIdentityProvider)
	object.ca = d.CA
	object.ldapAttributes, err = d.LDAPAttributes.unwrap()
	if err != nil {
		return
	}
	object.url = d.URL
	object.bindDN = d.BindDN
	object.bindPassword = d.BindPassword
	object.insecure = d.Insecure
	return
}
