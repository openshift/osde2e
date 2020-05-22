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

// MarshalClusterMetric writes a value of the 'cluster_metric' type to the given writer.
func MarshalClusterMetric(object *ClusterMetric, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeClusterMetric(object, stream)
	stream.Flush()
	return stream.Error
}

// writeClusterMetric writes a value of the 'cluster_metric' type to the given stream.
func writeClusterMetric(object *ClusterMetric, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if object.total != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("total")
		writeValue(object.total, stream)
		count++
	}
	if object.updatedTimestamp != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("updated_timestamp")
		stream.WriteString((*object.updatedTimestamp).Format(time.RFC3339))
		count++
	}
	if object.used != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("used")
		writeValue(object.used, stream)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalClusterMetric reads a value of the 'cluster_metric' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalClusterMetric(source interface{}) (object *ClusterMetric, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readClusterMetric(iterator)
	err = iterator.Error
	return
}

// readClusterMetric reads a value of the 'cluster_metric' type from the given iterator.
func readClusterMetric(iterator *jsoniter.Iterator) *ClusterMetric {
	object := &ClusterMetric{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "total":
			value := readValue(iterator)
			object.total = value
		case "updated_timestamp":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.updatedTimestamp = &value
		case "used":
			value := readValue(iterator)
			object.used = value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
