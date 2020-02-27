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

// permissionData is the data structure used internally to marshal and unmarshal
// objects of type 'permission'.
type permissionData struct {
	Kind         *string "json:\"kind,omitempty\""
	ID           *string "json:\"id,omitempty\""
	HREF         *string "json:\"href,omitempty\""
	Action       *Action "json:\"action,omitempty\""
	ResourceType *string "json:\"resource_type,omitempty\""
	RoleID       *string "json:\"role_id,omitempty\""
}

// MarshalPermission writes a value of the 'permission' to the given target,
// which can be a writer or a JSON encoder.
func MarshalPermission(object *Permission, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'permission'
// value to a JSON document.
func (o *Permission) wrap() (data *permissionData, err error) {
	if o == nil {
		return
	}
	data = new(permissionData)
	data.ID = o.id
	data.HREF = o.href
	data.Kind = new(string)
	if o.link {
		*data.Kind = PermissionLinkKind
	} else {
		*data.Kind = PermissionKind
	}
	data.Action = o.action
	data.ResourceType = o.resourceType
	data.RoleID = o.roleID
	return
}

// UnmarshalPermission reads a value of the 'permission' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalPermission(source interface{}) (object *Permission, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(permissionData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'permission' type.
func (d *permissionData) unwrap() (object *Permission, err error) {
	if d == nil {
		return
	}
	object = new(Permission)
	object.id = d.ID
	object.href = d.HREF
	if d.Kind != nil {
		object.link = *d.Kind == PermissionLinkKind
	}
	object.action = d.Action
	object.resourceType = d.ResourceType
	object.roleID = d.RoleID
	return
}
