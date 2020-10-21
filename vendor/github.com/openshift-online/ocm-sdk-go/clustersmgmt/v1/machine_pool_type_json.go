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
	"net/http"
	"sort"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalMachinePool writes a value of the 'machine_pool' type to the given writer.
func MarshalMachinePool(object *MachinePool, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeMachinePool(object, stream)
	stream.Flush()
	return stream.Error
}

// writeMachinePool writes a value of the 'machine_pool' type to the given stream.
func writeMachinePool(object *MachinePool, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(MachinePoolLinkKind)
	} else {
		stream.WriteString(MachinePoolKind)
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
	if object.availabilityZones != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("availability_zones")
		writeStringList(object.availabilityZones, stream)
		count++
	}
	if object.cluster != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cluster")
		writeCluster(object.cluster, stream)
		count++
	}
	if object.instanceType != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("instance_type")
		stream.WriteString(*object.instanceType)
		count++
	}
	if object.labels != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("labels")
		stream.WriteObjectStart()
		keys := make([]string, len(object.labels))
		i := 0
		for key := range object.labels {
			keys[i] = key
			i++
		}
		sort.Strings(keys)
		for i, key := range keys {
			if i > 0 {
				stream.WriteMore()
			}
			item := object.labels[key]
			stream.WriteObjectField(key)
			stream.WriteString(item)
		}
		stream.WriteObjectEnd()
		count++
	}
	if object.replicas != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("replicas")
		stream.WriteInt(*object.replicas)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalMachinePool reads a value of the 'machine_pool' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalMachinePool(source interface{}) (object *MachinePool, err error) {
	if source == http.NoBody {
		return
	}
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readMachinePool(iterator)
	err = iterator.Error
	return
}

// readMachinePool reads a value of the 'machine_pool' type from the given iterator.
func readMachinePool(iterator *jsoniter.Iterator) *MachinePool {
	object := &MachinePool{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == MachinePoolLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "availability_zones":
			value := readStringList(iterator)
			object.availabilityZones = value
		case "cluster":
			value := readCluster(iterator)
			object.cluster = value
		case "instance_type":
			value := iterator.ReadString()
			object.instanceType = &value
		case "labels":
			value := map[string]string{}
			for {
				key := iterator.ReadObject()
				if key == "" {
					break
				}
				item := iterator.ReadString()
				value[key] = item
			}
			object.labels = value
		case "replicas":
			value := iterator.ReadInt()
			object.replicas = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
