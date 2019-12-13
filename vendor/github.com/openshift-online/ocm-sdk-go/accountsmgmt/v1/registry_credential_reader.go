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

import (
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// registryCredentialData is the data structure used internally to marshal and unmarshal
// objects of type 'registry_credential'.
type registryCredentialData struct {
	Kind     *string       "json:\"kind,omitempty\""
	ID       *string       "json:\"id,omitempty\""
	HREF     *string       "json:\"href,omitempty\""
	Account  *accountData  "json:\"account,omitempty\""
	Registry *registryData "json:\"registry,omitempty\""
	Token    *string       "json:\"token,omitempty\""
	Username *string       "json:\"username,omitempty\""
}

// MarshalRegistryCredential writes a value of the 'registry_credential' to the given target,
// which can be a writer or a JSON encoder.
func MarshalRegistryCredential(object *RegistryCredential, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'registry_credential'
// value to a JSON document.
func (o *RegistryCredential) wrap() (data *registryCredentialData, err error) {
	if o == nil {
		return
	}
	data = new(registryCredentialData)
	data.ID = o.id
	data.HREF = o.href
	data.Kind = new(string)
	if o.link {
		*data.Kind = RegistryCredentialLinkKind
	} else {
		*data.Kind = RegistryCredentialKind
	}
	data.Account, err = o.account.wrap()
	if err != nil {
		return
	}
	data.Registry, err = o.registry.wrap()
	if err != nil {
		return
	}
	data.Token = o.token
	data.Username = o.username
	return
}

// UnmarshalRegistryCredential reads a value of the 'registry_credential' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalRegistryCredential(source interface{}) (object *RegistryCredential, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(registryCredentialData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'registry_credential' type.
func (d *registryCredentialData) unwrap() (object *RegistryCredential, err error) {
	if d == nil {
		return
	}
	object = new(RegistryCredential)
	object.id = d.ID
	object.href = d.HREF
	if d.Kind != nil {
		object.link = *d.Kind == RegistryCredentialLinkKind
	}
	object.account, err = d.Account.unwrap()
	if err != nil {
		return
	}
	object.registry, err = d.Registry.unwrap()
	if err != nil {
		return
	}
	object.token = d.Token
	object.username = d.Username
	return
}
