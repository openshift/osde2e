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

// CPUTotalsNodeRoleOSMetricNodeBuilder contains the data and logic needed to build 'CPU_totals_node_role_OS_metric_node' objects.
//
// Representation of information from telemetry about the CPU capacity by node
// role and OS of a cluster.
type CPUTotalsNodeRoleOSMetricNodeBuilder struct {
	cpuTotals []*CPUTotalNodeRoleOSMetricNodeBuilder
}

// NewCPUTotalsNodeRoleOSMetricNode creates a new builder of 'CPU_totals_node_role_OS_metric_node' objects.
func NewCPUTotalsNodeRoleOSMetricNode() *CPUTotalsNodeRoleOSMetricNodeBuilder {
	return new(CPUTotalsNodeRoleOSMetricNodeBuilder)
}

// CPUTotals sets the value of the 'CPU_totals' attribute
// to the given values.
//
//
func (b *CPUTotalsNodeRoleOSMetricNodeBuilder) CPUTotals(values ...*CPUTotalNodeRoleOSMetricNodeBuilder) *CPUTotalsNodeRoleOSMetricNodeBuilder {
	b.cpuTotals = make([]*CPUTotalNodeRoleOSMetricNodeBuilder, len(values))
	copy(b.cpuTotals, values)
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *CPUTotalsNodeRoleOSMetricNodeBuilder) Copy(object *CPUTotalsNodeRoleOSMetricNode) *CPUTotalsNodeRoleOSMetricNodeBuilder {
	if object == nil {
		return b
	}
	if object.cpuTotals != nil && len(object.cpuTotals.items) > 0 {
		b.cpuTotals = make([]*CPUTotalNodeRoleOSMetricNodeBuilder, len(object.cpuTotals.items))
		for i, item := range object.cpuTotals.items {
			b.cpuTotals[i] = NewCPUTotalNodeRoleOSMetricNode().Copy(item)
		}
	} else {
		b.cpuTotals = nil
	}
	return b
}

// Build creates a 'CPU_totals_node_role_OS_metric_node' object using the configuration stored in the builder.
func (b *CPUTotalsNodeRoleOSMetricNodeBuilder) Build() (object *CPUTotalsNodeRoleOSMetricNode, err error) {
	object = new(CPUTotalsNodeRoleOSMetricNode)
	if b.cpuTotals != nil {
		object.cpuTotals = new(CPUTotalNodeRoleOSMetricNodeList)
		object.cpuTotals.items = make([]*CPUTotalNodeRoleOSMetricNode, len(b.cpuTotals))
		for i, item := range b.cpuTotals {
			object.cpuTotals.items[i], err = item.Build()
			if err != nil {
				return
			}
		}
	}
	return
}
