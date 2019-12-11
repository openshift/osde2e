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
	time "time"

	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// reservedResourceData is the data structure used internally to marshal and unmarshal
// objects of type 'reserved_resource'.
type reservedResourceData struct {
	BYOC                 *bool      "json:\"byoc,omitempty\""
	AvailabilityZoneType *string    "json:\"availability_zone_type,omitempty\""
	Count                *int       "json:\"count,omitempty\""
	CreatedAt            *time.Time "json:\"created_at,omitempty\""
	ResourceName         *string    "json:\"resource_name,omitempty\""
	ResourceType         *string    "json:\"resource_type,omitempty\""
	UpdatedAt            *time.Time "json:\"updated_at,omitempty\""
}

// MarshalReservedResource writes a value of the 'reserved_resource' to the given target,
// which can be a writer or a JSON encoder.
func MarshalReservedResource(object *ReservedResource, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'reserved_resource'
// value to a JSON document.
func (o *ReservedResource) wrap() (data *reservedResourceData, err error) {
	if o == nil {
		return
	}
	data = new(reservedResourceData)
	data.BYOC = o.byoc
	data.AvailabilityZoneType = o.availabilityZoneType
	data.Count = o.count
	data.CreatedAt = o.createdAt
	data.ResourceName = o.resourceName
	data.ResourceType = o.resourceType
	data.UpdatedAt = o.updatedAt
	return
}

// UnmarshalReservedResource reads a value of the 'reserved_resource' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalReservedResource(source interface{}) (object *ReservedResource, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(reservedResourceData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'reserved_resource' type.
func (d *reservedResourceData) unwrap() (object *ReservedResource, err error) {
	if d == nil {
		return
	}
	object = new(ReservedResource)
	object.byoc = d.BYOC
	object.availabilityZoneType = d.AvailabilityZoneType
	object.count = d.Count
	object.createdAt = d.CreatedAt
	object.resourceName = d.ResourceName
	object.resourceType = d.ResourceType
	object.updatedAt = d.UpdatedAt
	return
}
