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

package v1 // github.com/openshift-online/ocm-sdk-go/authorizations/v1

import (
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// exportControlReviewRequestData is the data structure used internally to marshal and unmarshal
// objects of type 'export_control_review_request'.
type exportControlReviewRequestData struct {
	AccountUsername *string "json:\"account_username,omitempty\""
}

// MarshalExportControlReviewRequest writes a value of the 'export_control_review_request' to the given target,
// which can be a writer or a JSON encoder.
func MarshalExportControlReviewRequest(object *ExportControlReviewRequest, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'export_control_review_request'
// value to a JSON document.
func (o *ExportControlReviewRequest) wrap() (data *exportControlReviewRequestData, err error) {
	if o == nil {
		return
	}
	data = new(exportControlReviewRequestData)
	data.AccountUsername = o.accountUsername
	return
}

// UnmarshalExportControlReviewRequest reads a value of the 'export_control_review_request' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalExportControlReviewRequest(source interface{}) (object *ExportControlReviewRequest, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(exportControlReviewRequestData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'export_control_review_request' type.
func (d *exportControlReviewRequestData) unwrap() (object *ExportControlReviewRequest, err error) {
	if d == nil {
		return
	}
	object = new(ExportControlReviewRequest)
	object.accountUsername = d.AccountUsername
	return
}
