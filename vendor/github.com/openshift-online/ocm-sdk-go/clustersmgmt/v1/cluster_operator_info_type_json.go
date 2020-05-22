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
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalClusterOperatorInfo writes a value of the 'cluster_operator_info' type to the given writer.
func MarshalClusterOperatorInfo(object *ClusterOperatorInfo, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeClusterOperatorInfo(object, stream)
	stream.Flush()
	return stream.Error
}

// writeClusterOperatorInfo writes a value of the 'cluster_operator_info' type to the given stream.
func writeClusterOperatorInfo(object *ClusterOperatorInfo, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if object.condition != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("condition")
		stream.WriteString(string(*object.condition))
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
	if object.reason != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("reason")
		stream.WriteString(*object.reason)
		count++
	}
	if object.time != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("time")
		stream.WriteString((*object.time).Format(time.RFC3339))
		count++
	}
	if object.version != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("version")
		stream.WriteString(*object.version)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalClusterOperatorInfo reads a value of the 'cluster_operator_info' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalClusterOperatorInfo(source interface{}) (object *ClusterOperatorInfo, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readClusterOperatorInfo(iterator)
	err = iterator.Error
	return
}

// readClusterOperatorInfo reads a value of the 'cluster_operator_info' type from the given iterator.
func readClusterOperatorInfo(iterator *jsoniter.Iterator) *ClusterOperatorInfo {
	object := &ClusterOperatorInfo{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "condition":
			text := iterator.ReadString()
			value := ClusterOperatorState(text)
			object.condition = &value
		case "name":
			value := iterator.ReadString()
			object.name = &value
		case "reason":
			value := iterator.ReadString()
			object.reason = &value
		case "time":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.time = &value
		case "version":
			value := iterator.ReadString()
			object.version = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
