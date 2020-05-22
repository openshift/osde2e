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

// MarshalClusterNodes writes a value of the 'cluster_nodes' type to the given writer.
func MarshalClusterNodes(object *ClusterNodes, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeClusterNodes(object, stream)
	stream.Flush()
	return stream.Error
}

// writeClusterNodes writes a value of the 'cluster_nodes' type to the given stream.
func writeClusterNodes(object *ClusterNodes, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if object.compute != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("compute")
		stream.WriteInt(*object.compute)
		count++
	}
	if object.computeMachineType != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("compute_machine_type")
		writeMachineType(object.computeMachineType, stream)
		count++
	}
	if object.infra != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("infra")
		stream.WriteInt(*object.infra)
		count++
	}
	if object.master != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("master")
		stream.WriteInt(*object.master)
		count++
	}
	if object.total != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("total")
		stream.WriteInt(*object.total)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalClusterNodes reads a value of the 'cluster_nodes' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalClusterNodes(source interface{}) (object *ClusterNodes, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readClusterNodes(iterator)
	err = iterator.Error
	return
}

// readClusterNodes reads a value of the 'cluster_nodes' type from the given iterator.
func readClusterNodes(iterator *jsoniter.Iterator) *ClusterNodes {
	object := &ClusterNodes{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "compute":
			value := iterator.ReadInt()
			object.compute = &value
		case "compute_machine_type":
			value := readMachineType(iterator)
			object.computeMachineType = value
		case "infra":
			value := iterator.ReadInt()
			object.infra = &value
		case "master":
			value := iterator.ReadInt()
			object.master = &value
		case "total":
			value := iterator.ReadInt()
			object.total = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
