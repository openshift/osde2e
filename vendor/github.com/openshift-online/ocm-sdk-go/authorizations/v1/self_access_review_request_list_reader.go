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

// selfAccessReviewRequestListData is type used internally to marshal and unmarshal lists of objects
// of type 'self_access_review_request'.
type selfAccessReviewRequestListData []*selfAccessReviewRequestData

// UnmarshalSelfAccessReviewRequestList reads a list of values of the 'self_access_review_request'
// from the given source, which can be a slice of bytes, a string, an io.Reader or a
// json.Decoder.
func UnmarshalSelfAccessReviewRequestList(source interface{}) (list *SelfAccessReviewRequestList, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	var data selfAccessReviewRequestListData
	err = decoder.Decode(&data)
	if err != nil {
		return
	}
	list, err = data.unwrap()
	return
}

// wrap is the method used internally to convert a list of values of the
// 'self_access_review_request' value to a JSON document.
func (l *SelfAccessReviewRequestList) wrap() (data selfAccessReviewRequestListData, err error) {
	if l == nil {
		return
	}
	data = make(selfAccessReviewRequestListData, len(l.items))
	for i, item := range l.items {
		data[i], err = item.wrap()
		if err != nil {
			return
		}
	}
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// list of values of the 'self_access_review_request' type.
func (d selfAccessReviewRequestListData) unwrap() (list *SelfAccessReviewRequestList, err error) {
	if d == nil {
		return
	}
	items := make([]*SelfAccessReviewRequest, len(d))
	for i, item := range d {
		items[i], err = item.unwrap()
		if err != nil {
			return
		}
	}
	list = new(SelfAccessReviewRequestList)
	list.items = items
	return
}
