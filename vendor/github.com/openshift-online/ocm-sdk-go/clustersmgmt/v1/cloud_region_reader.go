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

// cloudRegionData is the data structure used internally to marshal and unmarshal
// objects of type 'cloud_region'.
type cloudRegionData struct {
	Kind          *string            "json:\"kind,omitempty\""
	ID            *string            "json:\"id,omitempty\""
	HREF          *string            "json:\"href,omitempty\""
	CloudProvider *cloudProviderData "json:\"cloud_provider,omitempty\""
	DisplayName   *string            "json:\"display_name,omitempty\""
	Name          *string            "json:\"name,omitempty\""
}

// MarshalCloudRegion writes a value of the 'cloud_region' to the given target,
// which can be a writer or a JSON encoder.
func MarshalCloudRegion(object *CloudRegion, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'cloud_region'
// value to a JSON document.
func (o *CloudRegion) wrap() (data *cloudRegionData, err error) {
	if o == nil {
		return
	}
	data = new(cloudRegionData)
	data.ID = o.id
	data.HREF = o.href
	data.Kind = new(string)
	if o.link {
		*data.Kind = CloudRegionLinkKind
	} else {
		*data.Kind = CloudRegionKind
	}
	data.CloudProvider, err = o.cloudProvider.wrap()
	if err != nil {
		return
	}
	data.DisplayName = o.displayName
	data.Name = o.name
	return
}

// UnmarshalCloudRegion reads a value of the 'cloud_region' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalCloudRegion(source interface{}) (object *CloudRegion, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(cloudRegionData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'cloud_region' type.
func (d *cloudRegionData) unwrap() (object *CloudRegion, err error) {
	if d == nil {
		return
	}
	object = new(CloudRegion)
	object.id = d.ID
	object.href = d.HREF
	if d.Kind != nil {
		object.link = *d.Kind == CloudRegionLinkKind
	}
	object.cloudProvider, err = d.CloudProvider.unwrap()
	if err != nil {
		return
	}
	object.displayName = d.DisplayName
	object.name = d.Name
	return
}
