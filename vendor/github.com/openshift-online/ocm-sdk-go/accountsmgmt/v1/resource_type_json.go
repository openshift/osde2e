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
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalResource writes a value of the 'resource' type to the given writer.
func MarshalResource(object *Resource, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeResource(object, stream)
	stream.Flush()
	return stream.Error
}

// writeResource writes a value of the 'resource' type to the given stream.
func writeResource(object *Resource, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(ResourceLinkKind)
	} else {
		stream.WriteString(ResourceKind)
	}
	count++
	if object.id != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("id")
		stream.WriteString(*object.id)
		count++
	}
	if object.href != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("href")
		stream.WriteString(*object.href)
		count++
	}
	if object.byoc != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("byoc")
		stream.WriteBool(*object.byoc)
		count++
	}
	if object.sku != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("sku")
		stream.WriteString(*object.sku)
		count++
	}
	if object.allowed != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("allowed")
		stream.WriteInt(*object.allowed)
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
	stream.WriteObjectEnd()
}

// UnmarshalResource reads a value of the 'resource' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalResource(source interface{}) (object *Resource, err error) {
	if source == http.NoBody {
		return
	}
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readResource(iterator)
	err = iterator.Error
	return
}

// readResource reads a value of the 'resource' type from the given iterator.
func readResource(iterator *jsoniter.Iterator) *Resource {
	object := &Resource{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == ResourceLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "byoc":
			value := iterator.ReadBool()
			object.byoc = &value
		case "sku":
			value := iterator.ReadString()
			object.sku = &value
		case "allowed":
			value := iterator.ReadInt()
			object.allowed = &value
		case "availability_zone_type":
			value := iterator.ReadString()
			object.availabilityZoneType = &value
		case "resource_name":
			value := iterator.ReadString()
			object.resourceName = &value
		case "resource_type":
			value := iterator.ReadString()
			object.resourceType = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
