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

// clusterRegistrationResponseData is the data structure used internally to marshal and unmarshal
// objects of type 'cluster_registration_response'.
type clusterRegistrationResponseData struct {
	AccountID          *string "json:\"account_id,omitempty\""
	AuthorizationToken *string "json:\"authorization_token,omitempty\""
	ClusterID          *string "json:\"cluster_id,omitempty\""
	ExpiresAt          *string "json:\"expires_at,omitempty\""
}

// MarshalClusterRegistrationResponse writes a value of the 'cluster_registration_response' to the given target,
// which can be a writer or a JSON encoder.
func MarshalClusterRegistrationResponse(object *ClusterRegistrationResponse, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'cluster_registration_response'
// value to a JSON document.
func (o *ClusterRegistrationResponse) wrap() (data *clusterRegistrationResponseData, err error) {
	if o == nil {
		return
	}
	data = new(clusterRegistrationResponseData)
	data.AccountID = o.accountID
	data.AuthorizationToken = o.authorizationToken
	data.ClusterID = o.clusterID
	data.ExpiresAt = o.expiresAt
	return
}

// UnmarshalClusterRegistrationResponse reads a value of the 'cluster_registration_response' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalClusterRegistrationResponse(source interface{}) (object *ClusterRegistrationResponse, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(clusterRegistrationResponseData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'cluster_registration_response' type.
func (d *clusterRegistrationResponseData) unwrap() (object *ClusterRegistrationResponse, err error) {
	if d == nil {
		return
	}
	object = new(ClusterRegistrationResponse)
	object.accountID = d.AccountID
	object.authorizationToken = d.AuthorizationToken
	object.clusterID = d.ClusterID
	object.expiresAt = d.ExpiresAt
	return
}
