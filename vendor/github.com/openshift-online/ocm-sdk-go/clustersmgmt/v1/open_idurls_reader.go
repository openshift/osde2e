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

// openIDURLsData is the data structure used internally to marshal and unmarshal
// objects of type 'open_IDURLs'.
type openIDURLsData struct {
	Authorize *string "json:\"authorize,omitempty\""
	Token     *string "json:\"token,omitempty\""
	UserInfo  *string "json:\"user_info,omitempty\""
}

// MarshalOpenIDURLs writes a value of the 'open_IDURLs' to the given target,
// which can be a writer or a JSON encoder.
func MarshalOpenIDURLs(object *OpenIDURLs, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'open_IDURLs'
// value to a JSON document.
func (o *OpenIDURLs) wrap() (data *openIDURLsData, err error) {
	if o == nil {
		return
	}
	data = new(openIDURLsData)
	data.Authorize = o.authorize
	data.Token = o.token
	data.UserInfo = o.userInfo
	return
}

// UnmarshalOpenIDURLs reads a value of the 'open_IDURLs' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalOpenIDURLs(source interface{}) (object *OpenIDURLs, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(openIDURLsData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'open_IDURLs' type.
func (d *openIDURLsData) unwrap() (object *OpenIDURLs, err error) {
	if d == nil {
		return
	}
	object = new(OpenIDURLs)
	object.authorize = d.Authorize
	object.token = d.Token
	object.userInfo = d.UserInfo
	return
}
