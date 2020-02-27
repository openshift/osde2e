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

package v1 // github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1

import (
	time "time"

	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// cpuTotalNodeRoleOSMetricNodeData is the data structure used internally to marshal and unmarshal
// objects of type 'CPU_total_node_role_OS_metric_node'.
type cpuTotalNodeRoleOSMetricNodeData struct {
	CPUTotal        *float64   "json:\"cpu_total,omitempty\""
	NodeRoles       []string   "json:\"node_roles,omitempty\""
	OperatingSystem *string    "json:\"operating_system,omitempty\""
	Time            *time.Time "json:\"time,omitempty\""
}

// MarshalCPUTotalNodeRoleOSMetricNode writes a value of the 'CPU_total_node_role_OS_metric_node' to the given target,
// which can be a writer or a JSON encoder.
func MarshalCPUTotalNodeRoleOSMetricNode(object *CPUTotalNodeRoleOSMetricNode, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'CPU_total_node_role_OS_metric_node'
// value to a JSON document.
func (o *CPUTotalNodeRoleOSMetricNode) wrap() (data *cpuTotalNodeRoleOSMetricNodeData, err error) {
	if o == nil {
		return
	}
	data = new(cpuTotalNodeRoleOSMetricNodeData)
	data.CPUTotal = o.cpuTotal
	data.NodeRoles = o.nodeRoles
	data.OperatingSystem = o.operatingSystem
	data.Time = o.time
	return
}

// UnmarshalCPUTotalNodeRoleOSMetricNode reads a value of the 'CPU_total_node_role_OS_metric_node' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalCPUTotalNodeRoleOSMetricNode(source interface{}) (object *CPUTotalNodeRoleOSMetricNode, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(cpuTotalNodeRoleOSMetricNodeData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'CPU_total_node_role_OS_metric_node' type.
func (d *cpuTotalNodeRoleOSMetricNodeData) unwrap() (object *CPUTotalNodeRoleOSMetricNode, err error) {
	if d == nil {
		return
	}
	object = new(CPUTotalNodeRoleOSMetricNode)
	object.cpuTotal = d.CPUTotal
	object.nodeRoles = d.NodeRoles
	object.operatingSystem = d.OperatingSystem
	object.time = d.Time
	return
}
