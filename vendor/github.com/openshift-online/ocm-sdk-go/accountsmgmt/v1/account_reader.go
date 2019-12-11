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
	"fmt"

	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// accountData is the data structure used internally to marshal and unmarshal
// objects of type 'account'.
type accountData struct {
	Kind           *string           "json:\"kind,omitempty\""
	ID             *string           "json:\"id,omitempty\""
	HREF           *string           "json:\"href,omitempty\""
	BanDescription *string           "json:\"ban_description,omitempty\""
	Banned         *bool             "json:\"banned,omitempty\""
	Email          *string           "json:\"email,omitempty\""
	FirstName      *string           "json:\"first_name,omitempty\""
	LastName       *string           "json:\"last_name,omitempty\""
	Name           *string           "json:\"name,omitempty\""
	Organization   *organizationData "json:\"organization,omitempty\""
	Username       *string           "json:\"username,omitempty\""
}

// MarshalAccount writes a value of the 'account' to the given target,
// which can be a writer or a JSON encoder.
func MarshalAccount(object *Account, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'account'
// value to a JSON document.
func (o *Account) wrap() (data *accountData, err error) {
	if o == nil {
		return
	}
	data = new(accountData)
	data.ID = o.id
	data.HREF = o.href
	data.Kind = new(string)
	if o.link {
		*data.Kind = AccountLinkKind
	} else {
		*data.Kind = AccountKind
	}
	data.BanDescription = o.banDescription
	data.Banned = o.banned
	data.Email = o.email
	data.FirstName = o.firstName
	data.LastName = o.lastName
	data.Name = o.name
	data.Organization, err = o.organization.wrap()
	if err != nil {
		return
	}
	data.Username = o.username
	return
}

// UnmarshalAccount reads a value of the 'account' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalAccount(source interface{}) (object *Account, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(accountData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'account' type.
func (d *accountData) unwrap() (object *Account, err error) {
	if d == nil {
		return
	}
	object = new(Account)
	object.id = d.ID
	object.href = d.HREF
	if d.Kind != nil {
		switch *d.Kind {
		case AccountKind:
			object.link = false
		case AccountLinkKind:
			object.link = true
		default:
			err = fmt.Errorf(
				"expected kind '%s' or '%s' but got '%s'",
				AccountKind,
				AccountLinkKind,
				*d.Kind,
			)
			return
		}
	}
	object.banDescription = d.BanDescription
	object.banned = d.Banned
	object.email = d.Email
	object.firstName = d.FirstName
	object.lastName = d.LastName
	object.name = d.Name
	object.organization, err = d.Organization.unwrap()
	if err != nil {
		return
	}
	object.username = d.Username
	return
}
