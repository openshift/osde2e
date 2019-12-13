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

// identityProviderData is the data structure used internally to marshal and unmarshal
// objects of type 'identity_provider'.
type identityProviderData struct {
	Kind          *string                        "json:\"kind,omitempty\""
	ID            *string                        "json:\"id,omitempty\""
	HREF          *string                        "json:\"href,omitempty\""
	LDAP          *ldapIdentityProviderData      "json:\"ldap,omitempty\""
	Challenge     *bool                          "json:\"challenge,omitempty\""
	Github        *githubIdentityProviderData    "json:\"github,omitempty\""
	Gitlab        *gitlabIdentityProviderData    "json:\"gitlab,omitempty\""
	Google        *googleIdentityProviderData    "json:\"google,omitempty\""
	Login         *bool                          "json:\"login,omitempty\""
	MappingMethod *IdentityProviderMappingMethod "json:\"mapping_method,omitempty\""
	Name          *string                        "json:\"name,omitempty\""
	OpenID        *openIDIdentityProviderData    "json:\"open_id,omitempty\""
	Type          *IdentityProviderType          "json:\"type,omitempty\""
}

// MarshalIdentityProvider writes a value of the 'identity_provider' to the given target,
// which can be a writer or a JSON encoder.
func MarshalIdentityProvider(object *IdentityProvider, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'identity_provider'
// value to a JSON document.
func (o *IdentityProvider) wrap() (data *identityProviderData, err error) {
	if o == nil {
		return
	}
	data = new(identityProviderData)
	data.ID = o.id
	data.HREF = o.href
	data.Kind = new(string)
	if o.link {
		*data.Kind = IdentityProviderLinkKind
	} else {
		*data.Kind = IdentityProviderKind
	}
	data.LDAP, err = o.ldap.wrap()
	if err != nil {
		return
	}
	data.Challenge = o.challenge
	data.Github, err = o.github.wrap()
	if err != nil {
		return
	}
	data.Gitlab, err = o.gitlab.wrap()
	if err != nil {
		return
	}
	data.Google, err = o.google.wrap()
	if err != nil {
		return
	}
	data.Login = o.login
	data.MappingMethod = o.mappingMethod
	data.Name = o.name
	data.OpenID, err = o.openID.wrap()
	if err != nil {
		return
	}
	data.Type = o.type_
	return
}

// UnmarshalIdentityProvider reads a value of the 'identity_provider' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalIdentityProvider(source interface{}) (object *IdentityProvider, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(identityProviderData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'identity_provider' type.
func (d *identityProviderData) unwrap() (object *IdentityProvider, err error) {
	if d == nil {
		return
	}
	object = new(IdentityProvider)
	object.id = d.ID
	object.href = d.HREF
	if d.Kind != nil {
		object.link = *d.Kind == IdentityProviderLinkKind
	}
	object.ldap, err = d.LDAP.unwrap()
	if err != nil {
		return
	}
	object.challenge = d.Challenge
	object.github, err = d.Github.unwrap()
	if err != nil {
		return
	}
	object.gitlab, err = d.Gitlab.unwrap()
	if err != nil {
		return
	}
	object.google, err = d.Google.unwrap()
	if err != nil {
		return
	}
	object.login = d.Login
	object.mappingMethod = d.MappingMethod
	object.name = d.Name
	object.openID, err = d.OpenID.unwrap()
	if err != nil {
		return
	}
	object.type_ = d.Type
	return
}
