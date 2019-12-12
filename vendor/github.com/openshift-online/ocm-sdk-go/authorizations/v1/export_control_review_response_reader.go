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

// exportControlReviewResponseData is the data structure used internally to marshal and unmarshal
// objects of type 'export_control_review_response'.
type exportControlReviewResponseData struct {
	Restricted *bool "json:\"restricted,omitempty\""
}

// MarshalExportControlReviewResponse writes a value of the 'export_control_review_response' to the given target,
// which can be a writer or a JSON encoder.
func MarshalExportControlReviewResponse(object *ExportControlReviewResponse, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'export_control_review_response'
// value to a JSON document.
func (o *ExportControlReviewResponse) wrap() (data *exportControlReviewResponseData, err error) {
	if o == nil {
		return
	}
	data = new(exportControlReviewResponseData)
	data.Restricted = o.restricted
	return
}

// UnmarshalExportControlReviewResponse reads a value of the 'export_control_review_response' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalExportControlReviewResponse(source interface{}) (object *ExportControlReviewResponse, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(exportControlReviewResponseData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'export_control_review_response' type.
func (d *exportControlReviewResponseData) unwrap() (object *ExportControlReviewResponse, err error) {
	if d == nil {
		return
	}
	object = new(ExportControlReviewResponse)
	object.restricted = d.Restricted
	return
}
