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

// clusterMetricData is the data structure used internally to marshal and unmarshal
// objects of type 'cluster_metric'.
type clusterMetricData struct {
	Total            *valueData "json:\"total,omitempty\""
	UpdatedTimestamp *time.Time "json:\"updated_timestamp,omitempty\""
	Used             *valueData "json:\"used,omitempty\""
}

// MarshalClusterMetric writes a value of the 'cluster_metric' to the given target,
// which can be a writer or a JSON encoder.
func MarshalClusterMetric(object *ClusterMetric, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'cluster_metric'
// value to a JSON document.
func (o *ClusterMetric) wrap() (data *clusterMetricData, err error) {
	if o == nil {
		return
	}
	data = new(clusterMetricData)
	data.Total, err = o.total.wrap()
	if err != nil {
		return
	}
	data.UpdatedTimestamp = o.updatedTimestamp
	data.Used, err = o.used.wrap()
	if err != nil {
		return
	}
	return
}

// UnmarshalClusterMetric reads a value of the 'cluster_metric' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalClusterMetric(source interface{}) (object *ClusterMetric, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(clusterMetricData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'cluster_metric' type.
func (d *clusterMetricData) unwrap() (object *ClusterMetric, err error) {
	if d == nil {
		return
	}
	object = new(ClusterMetric)
	object.total, err = d.Total.unwrap()
	if err != nil {
		return
	}
	object.updatedTimestamp = d.UpdatedTimestamp
	object.used, err = d.Used.unwrap()
	if err != nil {
		return
	}
	return
}
