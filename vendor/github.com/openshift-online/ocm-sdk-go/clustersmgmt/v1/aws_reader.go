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

// awsData is the data structure used internally to marshal and unmarshal
// objects of type 'AWS'.
type awsData struct {
	AccessKeyID     *string "json:\"access_key_id,omitempty\""
	AccountID       *string "json:\"account_id,omitempty\""
	SecretAccessKey *string "json:\"secret_access_key,omitempty\""
}

// MarshalAWS writes a value of the 'AWS' to the given target,
// which can be a writer or a JSON encoder.
func MarshalAWS(object *AWS, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'AWS'
// value to a JSON document.
func (o *AWS) wrap() (data *awsData, err error) {
	if o == nil {
		return
	}
	data = new(awsData)
	data.AccessKeyID = o.accessKeyID
	data.AccountID = o.accountID
	data.SecretAccessKey = o.secretAccessKey
	return
}

// UnmarshalAWS reads a value of the 'AWS' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalAWS(source interface{}) (object *AWS, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(awsData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'AWS' type.
func (d *awsData) unwrap() (object *AWS, err error) {
	if d == nil {
		return
	}
	object = new(AWS)
	object.accessKeyID = d.AccessKeyID
	object.accountID = d.AccountID
	object.secretAccessKey = d.SecretAccessKey
	return
}
