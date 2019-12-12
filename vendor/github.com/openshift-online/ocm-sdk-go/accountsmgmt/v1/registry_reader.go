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

// registryData is the data structure used internally to marshal and unmarshal
// objects of type 'registry'.
type registryData struct {
	Kind       *string "json:\"kind,omitempty\""
	ID         *string "json:\"id,omitempty\""
	HREF       *string "json:\"href,omitempty\""
	URL        *string "json:\"url,omitempty\""
	CloudAlias *bool   "json:\"cloud_alias,omitempty\""
	Name       *string "json:\"name,omitempty\""
	OrgName    *string "json:\"org_name,omitempty\""
	TeamName   *string "json:\"team_name,omitempty\""
	Type       *string "json:\"type,omitempty\""
}

// MarshalRegistry writes a value of the 'registry' to the given target,
// which can be a writer or a JSON encoder.
func MarshalRegistry(object *Registry, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'registry'
// value to a JSON document.
func (o *Registry) wrap() (data *registryData, err error) {
	if o == nil {
		return
	}
	data = new(registryData)
	data.ID = o.id
	data.HREF = o.href
	data.Kind = new(string)
	if o.link {
		*data.Kind = RegistryLinkKind
	} else {
		*data.Kind = RegistryKind
	}
	data.URL = o.url
	data.CloudAlias = o.cloudAlias
	data.Name = o.name
	data.OrgName = o.orgName
	data.TeamName = o.teamName
	data.Type = o.type_
	return
}

// UnmarshalRegistry reads a value of the 'registry' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalRegistry(source interface{}) (object *Registry, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(registryData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'registry' type.
func (d *registryData) unwrap() (object *Registry, err error) {
	if d == nil {
		return
	}
	object = new(Registry)
	object.id = d.ID
	object.href = d.HREF
	if d.Kind != nil {
		switch *d.Kind {
		case RegistryKind:
			object.link = false
		case RegistryLinkKind:
			object.link = true
		default:
			err = fmt.Errorf(
				"expected kind '%s' or '%s' but got '%s'",
				RegistryKind,
				RegistryLinkKind,
				*d.Kind,
			)
			return
		}
	}
	object.url = d.URL
	object.cloudAlias = d.CloudAlias
	object.name = d.Name
	object.orgName = d.OrgName
	object.teamName = d.TeamName
	object.type_ = d.Type
	return
}
