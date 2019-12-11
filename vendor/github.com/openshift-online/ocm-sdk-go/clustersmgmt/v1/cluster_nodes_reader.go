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

// clusterNodesData is the data structure used internally to marshal and unmarshal
// objects of type 'cluster_nodes'.
type clusterNodesData struct {
	Compute *int "json:\"compute,omitempty\""
	Infra   *int "json:\"infra,omitempty\""
	Master  *int "json:\"master,omitempty\""
	Total   *int "json:\"total,omitempty\""
}

// MarshalClusterNodes writes a value of the 'cluster_nodes' to the given target,
// which can be a writer or a JSON encoder.
func MarshalClusterNodes(object *ClusterNodes, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'cluster_nodes'
// value to a JSON document.
func (o *ClusterNodes) wrap() (data *clusterNodesData, err error) {
	if o == nil {
		return
	}
	data = new(clusterNodesData)
	data.Compute = o.compute
	data.Infra = o.infra
	data.Master = o.master
	data.Total = o.total
	return
}

// UnmarshalClusterNodes reads a value of the 'cluster_nodes' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalClusterNodes(source interface{}) (object *ClusterNodes, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(clusterNodesData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'cluster_nodes' type.
func (d *clusterNodesData) unwrap() (object *ClusterNodes, err error) {
	if d == nil {
		return
	}
	object = new(ClusterNodes)
	object.compute = d.Compute
	object.infra = d.Infra
	object.master = d.Master
	object.total = d.Total
	return
}
