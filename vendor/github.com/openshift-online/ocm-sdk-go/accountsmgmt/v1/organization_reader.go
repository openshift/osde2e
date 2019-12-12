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

// organizationData is the data structure used internally to marshal and unmarshal
// objects of type 'organization'.
type organizationData struct {
	Kind       *string "json:\"kind,omitempty\""
	ID         *string "json:\"id,omitempty\""
	HREF       *string "json:\"href,omitempty\""
	ExternalID *string "json:\"external_id,omitempty\""
	Name       *string "json:\"name,omitempty\""
}

// MarshalOrganization writes a value of the 'organization' to the given target,
// which can be a writer or a JSON encoder.
func MarshalOrganization(object *Organization, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'organization'
// value to a JSON document.
func (o *Organization) wrap() (data *organizationData, err error) {
	if o == nil {
		return
	}
	data = new(organizationData)
	data.ID = o.id
	data.HREF = o.href
	data.Kind = new(string)
	if o.link {
		*data.Kind = OrganizationLinkKind
	} else {
		*data.Kind = OrganizationKind
	}
	data.ExternalID = o.externalID
	data.Name = o.name
	return
}

// UnmarshalOrganization reads a value of the 'organization' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalOrganization(source interface{}) (object *Organization, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(organizationData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'organization' type.
func (d *organizationData) unwrap() (object *Organization, err error) {
	if d == nil {
		return
	}
	object = new(Organization)
	object.id = d.ID
	object.href = d.HREF
	if d.Kind != nil {
		switch *d.Kind {
		case OrganizationKind:
			object.link = false
		case OrganizationLinkKind:
			object.link = true
		default:
			err = fmt.Errorf(
				"expected kind '%s' or '%s' but got '%s'",
				OrganizationKind,
				OrganizationLinkKind,
				*d.Kind,
			)
			return
		}
	}
	object.externalID = d.ExternalID
	object.name = d.Name
	return
}
