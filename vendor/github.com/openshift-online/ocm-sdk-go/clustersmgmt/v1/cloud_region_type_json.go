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

package v1 // github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1

import (
	"io"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalCloudRegion writes a value of the 'cloud_region' type to the given writer.
func MarshalCloudRegion(object *CloudRegion, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeCloudRegion(object, stream)
	stream.Flush()
	return stream.Error
}

// writeCloudRegion writes a value of the 'cloud_region' type to the given stream.
func writeCloudRegion(object *CloudRegion, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(CloudRegionLinkKind)
	} else {
		stream.WriteString(CloudRegionKind)
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
	if object.cloudProvider != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cloud_provider")
		writeCloudProvider(object.cloudProvider, stream)
		count++
	}
	if object.displayName != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("display_name")
		stream.WriteString(*object.displayName)
		count++
	}
	if object.enabled != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("enabled")
		stream.WriteBool(*object.enabled)
		count++
	}
	if object.name != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("name")
		stream.WriteString(*object.name)
		count++
	}
	if object.supportsMultiAZ != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("supports_multi_az")
		stream.WriteBool(*object.supportsMultiAZ)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalCloudRegion reads a value of the 'cloud_region' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalCloudRegion(source interface{}) (object *CloudRegion, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readCloudRegion(iterator)
	err = iterator.Error
	return
}

// readCloudRegion reads a value of the 'cloud_region' type from the given iterator.
func readCloudRegion(iterator *jsoniter.Iterator) *CloudRegion {
	object := &CloudRegion{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == CloudRegionLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "cloud_provider":
			value := readCloudProvider(iterator)
			object.cloudProvider = value
		case "display_name":
			value := iterator.ReadString()
			object.displayName = &value
		case "enabled":
			value := iterator.ReadBool()
			object.enabled = &value
		case "name":
			value := iterator.ReadString()
			object.name = &value
		case "supports_multi_az":
			value := iterator.ReadBool()
			object.supportsMultiAZ = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
