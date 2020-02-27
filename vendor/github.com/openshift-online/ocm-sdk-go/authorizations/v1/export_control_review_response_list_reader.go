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

// exportControlReviewResponseListData is type used internally to marshal and unmarshal lists of objects
// of type 'export_control_review_response'.
type exportControlReviewResponseListData []*exportControlReviewResponseData

// UnmarshalExportControlReviewResponseList reads a list of values of the 'export_control_review_response'
// from the given source, which can be a slice of bytes, a string, an io.Reader or a
// json.Decoder.
func UnmarshalExportControlReviewResponseList(source interface{}) (list *ExportControlReviewResponseList, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	var data exportControlReviewResponseListData
	err = decoder.Decode(&data)
	if err != nil {
		return
	}
	list, err = data.unwrap()
	return
}

// wrap is the method used internally to convert a list of values of the
// 'export_control_review_response' value to a JSON document.
func (l *ExportControlReviewResponseList) wrap() (data exportControlReviewResponseListData, err error) {
	if l == nil {
		return
	}
	data = make(exportControlReviewResponseListData, len(l.items))
	for i, item := range l.items {
		data[i], err = item.wrap()
		if err != nil {
			return
		}
	}
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// list of values of the 'export_control_review_response' type.
func (d exportControlReviewResponseListData) unwrap() (list *ExportControlReviewResponseList, err error) {
	if d == nil {
		return
	}
	items := make([]*ExportControlReviewResponse, len(d))
	for i, item := range d {
		items[i], err = item.unwrap()
		if err != nil {
			return
		}
	}
	list = new(ExportControlReviewResponseList)
	list.items = items
	return
}
