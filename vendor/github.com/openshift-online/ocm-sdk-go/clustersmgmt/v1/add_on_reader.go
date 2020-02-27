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

// addOnData is the data structure used internally to marshal and unmarshal
// objects of type 'add_on'.
type addOnData struct {
	Kind         *string  "json:\"kind,omitempty\""
	ID           *string  "json:\"id,omitempty\""
	HREF         *string  "json:\"href,omitempty\""
	Description  *string  "json:\"description,omitempty\""
	Enabled      *bool    "json:\"enabled,omitempty\""
	Icon         *string  "json:\"icon,omitempty\""
	Label        *string  "json:\"label,omitempty\""
	Name         *string  "json:\"name,omitempty\""
	ResourceCost *float64 "json:\"resource_cost,omitempty\""
	ResourceName *string  "json:\"resource_name,omitempty\""
}

// MarshalAddOn writes a value of the 'add_on' to the given target,
// which can be a writer or a JSON encoder.
func MarshalAddOn(object *AddOn, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'add_on'
// value to a JSON document.
func (o *AddOn) wrap() (data *addOnData, err error) {
	if o == nil {
		return
	}
	data = new(addOnData)
	data.ID = o.id
	data.HREF = o.href
	data.Kind = new(string)
	if o.link {
		*data.Kind = AddOnLinkKind
	} else {
		*data.Kind = AddOnKind
	}
	data.Description = o.description
	data.Enabled = o.enabled
	data.Icon = o.icon
	data.Label = o.label
	data.Name = o.name
	data.ResourceCost = o.resourceCost
	data.ResourceName = o.resourceName
	return
}

// UnmarshalAddOn reads a value of the 'add_on' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalAddOn(source interface{}) (object *AddOn, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(addOnData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'add_on' type.
func (d *addOnData) unwrap() (object *AddOn, err error) {
	if d == nil {
		return
	}
	object = new(AddOn)
	object.id = d.ID
	object.href = d.HREF
	if d.Kind != nil {
		object.link = *d.Kind == AddOnLinkKind
	}
	object.description = d.Description
	object.enabled = d.Enabled
	object.icon = d.Icon
	object.label = d.Label
	object.name = d.Name
	object.resourceCost = d.ResourceCost
	object.resourceName = d.ResourceName
	return
}
