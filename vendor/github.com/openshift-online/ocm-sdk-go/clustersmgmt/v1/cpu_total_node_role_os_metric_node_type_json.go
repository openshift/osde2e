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

// MarshalCPUTotalNodeRoleOSMetricNode writes a value of the 'CPU_total_node_role_OS_metric_node' type to the given writer.
func MarshalCPUTotalNodeRoleOSMetricNode(object *CPUTotalNodeRoleOSMetricNode, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeCPUTotalNodeRoleOSMetricNode(object, stream)
	stream.Flush()
	return stream.Error
}

// writeCPUTotalNodeRoleOSMetricNode writes a value of the 'CPU_total_node_role_OS_metric_node' type to the given stream.
func writeCPUTotalNodeRoleOSMetricNode(object *CPUTotalNodeRoleOSMetricNode, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if object.cpuTotal != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cpu_total")
		stream.WriteFloat64(*object.cpuTotal)
		count++
	}
	if object.nodeRoles != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("node_roles")
		writeStringList(object.nodeRoles, stream)
		count++
	}
	if object.operatingSystem != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("operating_system")
		stream.WriteString(*object.operatingSystem)
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
	stream.WriteObjectEnd()
}

// UnmarshalCPUTotalNodeRoleOSMetricNode reads a value of the 'CPU_total_node_role_OS_metric_node' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalCPUTotalNodeRoleOSMetricNode(source interface{}) (object *CPUTotalNodeRoleOSMetricNode, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readCPUTotalNodeRoleOSMetricNode(iterator)
	err = iterator.Error
	return
}

// readCPUTotalNodeRoleOSMetricNode reads a value of the 'CPU_total_node_role_OS_metric_node' type from the given iterator.
func readCPUTotalNodeRoleOSMetricNode(iterator *jsoniter.Iterator) *CPUTotalNodeRoleOSMetricNode {
	object := &CPUTotalNodeRoleOSMetricNode{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "cpu_total":
			value := iterator.ReadFloat64()
			object.cpuTotal = &value
		case "node_roles":
			value := readStringList(iterator)
			object.nodeRoles = value
		case "operating_system":
			value := iterator.ReadString()
			object.operatingSystem = &value
		case "time":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.time = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
