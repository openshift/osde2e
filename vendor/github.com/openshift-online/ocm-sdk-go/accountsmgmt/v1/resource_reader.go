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

// resourceData is the data structure used internally to marshal and unmarshal
// objects of type 'resource'.
type resourceData struct {
	Kind         *string "json:\"kind,omitempty\""
	ID           *string "json:\"id,omitempty\""
	HREF         *string "json:\"href,omitempty\""
	Allowed      *int    "json:\"allowed,omitempty\""
	ResourceName *string "json:\"resource_name,omitempty\""
	ResourceType *string "json:\"resource_type,omitempty\""
}

// MarshalResource writes a value of the 'resource' to the given target,
// which can be a writer or a JSON encoder.
func MarshalResource(object *Resource, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'resource'
// value to a JSON document.
func (o *Resource) wrap() (data *resourceData, err error) {
	if o == nil {
		return
	}
	data = new(resourceData)
	data.ID = o.id
	data.HREF = o.href
	data.Kind = new(string)
	if o.link {
		*data.Kind = ResourceLinkKind
	} else {
		*data.Kind = ResourceKind
	}
	data.Allowed = o.allowed
	data.ResourceName = o.resourceName
	data.ResourceType = o.resourceType
	return
}

// UnmarshalResource reads a value of the 'resource' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalResource(source interface{}) (object *Resource, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(resourceData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'resource' type.
func (d *resourceData) unwrap() (object *Resource, err error) {
	if d == nil {
		return
	}
	object = new(Resource)
	object.id = d.ID
	object.href = d.HREF
	if d.Kind != nil {
		object.link = *d.Kind == ResourceLinkKind
	}
	object.allowed = d.Allowed
	object.resourceName = d.ResourceName
	object.resourceType = d.ResourceType
	return
}
