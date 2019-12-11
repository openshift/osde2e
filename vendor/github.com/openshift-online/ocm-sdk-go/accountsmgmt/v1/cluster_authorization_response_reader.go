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

package v1 // github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1

import (
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// clusterAuthorizationResponseData is the data structure used internally to marshal and unmarshal
// objects of type 'cluster_authorization_response'.
type clusterAuthorizationResponseData struct {
	Allowed         *bool                    "json:\"allowed,omitempty\""
	ExcessResources reservedResourceListData "json:\"excess_resources,omitempty\""
	Subscription    *subscriptionData        "json:\"subscription,omitempty\""
}

// MarshalClusterAuthorizationResponse writes a value of the 'cluster_authorization_response' to the given target,
// which can be a writer or a JSON encoder.
func MarshalClusterAuthorizationResponse(object *ClusterAuthorizationResponse, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'cluster_authorization_response'
// value to a JSON document.
func (o *ClusterAuthorizationResponse) wrap() (data *clusterAuthorizationResponseData, err error) {
	if o == nil {
		return
	}
	data = new(clusterAuthorizationResponseData)
	data.Allowed = o.allowed
	data.ExcessResources, err = o.excessResources.wrap()
	if err != nil {
		return
	}
	data.Subscription, err = o.subscription.wrap()
	if err != nil {
		return
	}
	return
}

// UnmarshalClusterAuthorizationResponse reads a value of the 'cluster_authorization_response' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalClusterAuthorizationResponse(source interface{}) (object *ClusterAuthorizationResponse, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(clusterAuthorizationResponseData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'cluster_authorization_response' type.
func (d *clusterAuthorizationResponseData) unwrap() (object *ClusterAuthorizationResponse, err error) {
	if d == nil {
		return
	}
	object = new(ClusterAuthorizationResponse)
	object.allowed = d.Allowed
	object.excessResources, err = d.ExcessResources.unwrap()
	if err != nil {
		return
	}
	object.subscription, err = d.Subscription.unwrap()
	if err != nil {
		return
	}
	return
}
