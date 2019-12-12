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
	"fmt"

	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// clusterStatusData is the data structure used internally to marshal and unmarshal
// objects of type 'cluster_status'.
type clusterStatusData struct {
	Kind        *string       "json:\"kind,omitempty\""
	ID          *string       "json:\"id,omitempty\""
	HREF        *string       "json:\"href,omitempty\""
	Description *string       "json:\"description,omitempty\""
	State       *ClusterState "json:\"state,omitempty\""
}

// MarshalClusterStatus writes a value of the 'cluster_status' to the given target,
// which can be a writer or a JSON encoder.
func MarshalClusterStatus(object *ClusterStatus, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'cluster_status'
// value to a JSON document.
func (o *ClusterStatus) wrap() (data *clusterStatusData, err error) {
	if o == nil {
		return
	}
	data = new(clusterStatusData)
	data.ID = o.id
	data.HREF = o.href
	data.Kind = new(string)
	if o.link {
		*data.Kind = ClusterStatusLinkKind
	} else {
		*data.Kind = ClusterStatusKind
	}
	data.Description = o.description
	data.State = o.state
	return
}

// UnmarshalClusterStatus reads a value of the 'cluster_status' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalClusterStatus(source interface{}) (object *ClusterStatus, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(clusterStatusData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'cluster_status' type.
func (d *clusterStatusData) unwrap() (object *ClusterStatus, err error) {
	if d == nil {
		return
	}
	object = new(ClusterStatus)
	object.id = d.ID
	object.href = d.HREF
	if d.Kind != nil {
		switch *d.Kind {
		case ClusterStatusKind:
			object.link = false
		case ClusterStatusLinkKind:
			object.link = true
		default:
			err = fmt.Errorf(
				"expected kind '%s' or '%s' but got '%s'",
				ClusterStatusKind,
				ClusterStatusLinkKind,
				*d.Kind,
			)
			return
		}
	}
	object.description = d.Description
	object.state = d.State
	return
}
