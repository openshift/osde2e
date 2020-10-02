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

// MarshalMachineType writes a value of the 'machine_type' type to the given writer.
func MarshalMachineType(object *MachineType, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeMachineType(object, stream)
	stream.Flush()
	return stream.Error
}

// writeMachineType writes a value of the 'machine_type' type to the given stream.
func writeMachineType(object *MachineType, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(MachineTypeLinkKind)
	} else {
		stream.WriteString(MachineTypeKind)
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
	if object.cpu != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cpu")
		writeValue(object.cpu, stream)
		count++
	}
	if object.category != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("category")
		stream.WriteString(string(*object.category))
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
	if object.memory != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("memory")
		writeValue(object.memory, stream)
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
	if object.size != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("size")
		stream.WriteString(string(*object.size))
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalMachineType reads a value of the 'machine_type' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalMachineType(source interface{}) (object *MachineType, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readMachineType(iterator)
	err = iterator.Error
	return
}

// readMachineType reads a value of the 'machine_type' type from the given iterator.
func readMachineType(iterator *jsoniter.Iterator) *MachineType {
	object := &MachineType{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == MachineTypeLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "cpu":
			value := readValue(iterator)
			object.cpu = value
		case "category":
			text := iterator.ReadString()
			value := MachineTypeCategory(text)
			object.category = &value
		case "cloud_provider":
			value := readCloudProvider(iterator)
			object.cloudProvider = value
		case "memory":
			value := readValue(iterator)
			object.memory = value
		case "name":
			value := iterator.ReadString()
			object.name = &value
		case "size":
			text := iterator.ReadString()
			value := MachineTypeSize(text)
			object.size = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
