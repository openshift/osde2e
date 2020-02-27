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
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// clusterMetricsData is the data structure used internally to marshal and unmarshal
// objects of type 'cluster_metrics'.
type clusterMetricsData struct {
	CPU                *clusterMetricData "json:\"cpu,omitempty\""
	ComputeNodesCPU    *clusterMetricData "json:\"compute_nodes_cpu,omitempty\""
	ComputeNodesMemory *clusterMetricData "json:\"compute_nodes_memory,omitempty\""
	Memory             *clusterMetricData "json:\"memory,omitempty\""
	Nodes              *clusterNodesData  "json:\"nodes,omitempty\""
	Storage            *clusterMetricData "json:\"storage,omitempty\""
}

// MarshalClusterMetrics writes a value of the 'cluster_metrics' to the given target,
// which can be a writer or a JSON encoder.
func MarshalClusterMetrics(object *ClusterMetrics, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'cluster_metrics'
// value to a JSON document.
func (o *ClusterMetrics) wrap() (data *clusterMetricsData, err error) {
	if o == nil {
		return
	}
	data = new(clusterMetricsData)
	data.CPU, err = o.cpu.wrap()
	if err != nil {
		return
	}
	data.ComputeNodesCPU, err = o.computeNodesCPU.wrap()
	if err != nil {
		return
	}
	data.ComputeNodesMemory, err = o.computeNodesMemory.wrap()
	if err != nil {
		return
	}
	data.Memory, err = o.memory.wrap()
	if err != nil {
		return
	}
	data.Nodes, err = o.nodes.wrap()
	if err != nil {
		return
	}
	data.Storage, err = o.storage.wrap()
	if err != nil {
		return
	}
	return
}

// UnmarshalClusterMetrics reads a value of the 'cluster_metrics' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalClusterMetrics(source interface{}) (object *ClusterMetrics, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(clusterMetricsData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'cluster_metrics' type.
func (d *clusterMetricsData) unwrap() (object *ClusterMetrics, err error) {
	if d == nil {
		return
	}
	object = new(ClusterMetrics)
	object.cpu, err = d.CPU.unwrap()
	if err != nil {
		return
	}
	object.computeNodesCPU, err = d.ComputeNodesCPU.unwrap()
	if err != nil {
		return
	}
	object.computeNodesMemory, err = d.ComputeNodesMemory.unwrap()
	if err != nil {
		return
	}
	object.memory, err = d.Memory.unwrap()
	if err != nil {
		return
	}
	object.nodes, err = d.Nodes.unwrap()
	if err != nil {
		return
	}
	object.storage, err = d.Storage.unwrap()
	if err != nil {
		return
	}
	return
}
