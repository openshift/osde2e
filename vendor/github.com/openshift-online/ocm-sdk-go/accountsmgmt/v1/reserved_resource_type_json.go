/*
Copyright (c) 2020 Red Hat, Inc.

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
	"io"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalReservedResource writes a value of the 'reserved_resource' type to the given writer.
func MarshalReservedResource(object *ReservedResource, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeReservedResource(object, stream)
	stream.Flush()
	return stream.Error
}

// writeReservedResource writes a value of the 'reserved_resource' type to the given stream.
func writeReservedResource(object *ReservedResource, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if object.byoc != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("byoc")
		stream.WriteBool(*object.byoc)
		count++
	}
	if object.availabilityZoneType != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("availability_zone_type")
		stream.WriteString(*object.availabilityZoneType)
		count++
	}
	if object.count != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("count")
		stream.WriteInt(*object.count)
		count++
	}
	if object.createdAt != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("created_at")
		stream.WriteString((*object.createdAt).Format(time.RFC3339))
		count++
	}
	if object.resourceName != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("resource_name")
		stream.WriteString(*object.resourceName)
		count++
	}
	if object.resourceType != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("resource_type")
		stream.WriteString(*object.resourceType)
		count++
	}
	if object.updatedAt != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("updated_at")
		stream.WriteString((*object.updatedAt).Format(time.RFC3339))
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalReservedResource reads a value of the 'reserved_resource' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalReservedResource(source interface{}) (object *ReservedResource, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readReservedResource(iterator)
	err = iterator.Error
	return
}

// readReservedResource reads a value of the 'reserved_resource' type from the given iterator.
func readReservedResource(iterator *jsoniter.Iterator) *ReservedResource {
	object := &ReservedResource{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "byoc":
			value := iterator.ReadBool()
			object.byoc = &value
		case "availability_zone_type":
			value := iterator.ReadString()
			object.availabilityZoneType = &value
		case "count":
			value := iterator.ReadInt()
			object.count = &value
		case "created_at":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.createdAt = &value
		case "resource_name":
			value := iterator.ReadString()
			object.resourceName = &value
		case "resource_type":
			value := iterator.ReadString()
			object.resourceType = &value
		case "updated_at":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.updatedAt = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
