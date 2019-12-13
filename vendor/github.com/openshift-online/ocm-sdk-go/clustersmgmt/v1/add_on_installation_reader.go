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

// addOnInstallationData is the data structure used internally to marshal and unmarshal
// objects of type 'add_on_installation'.
type addOnInstallationData struct {
	Kind    *string      "json:\"kind,omitempty\""
	ID      *string      "json:\"id,omitempty\""
	HREF    *string      "json:\"href,omitempty\""
	Addon   *addOnData   "json:\"addon,omitempty\""
	Cluster *clusterData "json:\"cluster,omitempty\""
}

// MarshalAddOnInstallation writes a value of the 'add_on_installation' to the given target,
// which can be a writer or a JSON encoder.
func MarshalAddOnInstallation(object *AddOnInstallation, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'add_on_installation'
// value to a JSON document.
func (o *AddOnInstallation) wrap() (data *addOnInstallationData, err error) {
	if o == nil {
		return
	}
	data = new(addOnInstallationData)
	data.ID = o.id
	data.HREF = o.href
	data.Kind = new(string)
	if o.link {
		*data.Kind = AddOnInstallationLinkKind
	} else {
		*data.Kind = AddOnInstallationKind
	}
	data.Addon, err = o.addon.wrap()
	if err != nil {
		return
	}
	data.Cluster, err = o.cluster.wrap()
	if err != nil {
		return
	}
	return
}

// UnmarshalAddOnInstallation reads a value of the 'add_on_installation' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalAddOnInstallation(source interface{}) (object *AddOnInstallation, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(addOnInstallationData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'add_on_installation' type.
func (d *addOnInstallationData) unwrap() (object *AddOnInstallation, err error) {
	if d == nil {
		return
	}
	object = new(AddOnInstallation)
	object.id = d.ID
	object.href = d.HREF
	if d.Kind != nil {
		object.link = *d.Kind == AddOnInstallationLinkKind
	}
	object.addon, err = d.Addon.unwrap()
	if err != nil {
		return
	}
	object.cluster, err = d.Cluster.unwrap()
	if err != nil {
		return
	}
	return
}
