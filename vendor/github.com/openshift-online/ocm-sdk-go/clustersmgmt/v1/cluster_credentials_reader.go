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

// clusterCredentialsData is the data structure used internally to marshal and unmarshal
// objects of type 'cluster_credentials'.
type clusterCredentialsData struct {
	Kind       *string               "json:\"kind,omitempty\""
	ID         *string               "json:\"id,omitempty\""
	HREF       *string               "json:\"href,omitempty\""
	SSH        *sshCredentialsData   "json:\"ssh,omitempty\""
	Admin      *adminCredentialsData "json:\"admin,omitempty\""
	Kubeconfig *string               "json:\"kubeconfig,omitempty\""
}

// MarshalClusterCredentials writes a value of the 'cluster_credentials' to the given target,
// which can be a writer or a JSON encoder.
func MarshalClusterCredentials(object *ClusterCredentials, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'cluster_credentials'
// value to a JSON document.
func (o *ClusterCredentials) wrap() (data *clusterCredentialsData, err error) {
	if o == nil {
		return
	}
	data = new(clusterCredentialsData)
	data.ID = o.id
	data.HREF = o.href
	data.Kind = new(string)
	if o.link {
		*data.Kind = ClusterCredentialsLinkKind
	} else {
		*data.Kind = ClusterCredentialsKind
	}
	data.SSH, err = o.ssh.wrap()
	if err != nil {
		return
	}
	data.Admin, err = o.admin.wrap()
	if err != nil {
		return
	}
	data.Kubeconfig = o.kubeconfig
	return
}

// UnmarshalClusterCredentials reads a value of the 'cluster_credentials' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalClusterCredentials(source interface{}) (object *ClusterCredentials, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(clusterCredentialsData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'cluster_credentials' type.
func (d *clusterCredentialsData) unwrap() (object *ClusterCredentials, err error) {
	if d == nil {
		return
	}
	object = new(ClusterCredentials)
	object.id = d.ID
	object.href = d.HREF
	if d.Kind != nil {
		switch *d.Kind {
		case ClusterCredentialsKind:
			object.link = false
		case ClusterCredentialsLinkKind:
			object.link = true
		default:
			err = fmt.Errorf(
				"expected kind '%s' or '%s' but got '%s'",
				ClusterCredentialsKind,
				ClusterCredentialsLinkKind,
				*d.Kind,
			)
			return
		}
	}
	object.ssh, err = d.SSH.unwrap()
	if err != nil {
		return
	}
	object.admin, err = d.Admin.unwrap()
	if err != nil {
		return
	}
	object.kubeconfig = d.Kubeconfig
	return
}
