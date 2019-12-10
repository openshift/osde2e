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
	"fmt"

	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// planListData is type used internally to marshal and unmarshal lists of objects
// of type 'plan'.
type planListData []*planData

// UnmarshalPlanList reads a list of values of the 'plan'
// from the given source, which can be a slice of bytes, a string, an io.Reader or a
// json.Decoder.
func UnmarshalPlanList(source interface{}) (list *PlanList, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	var data planListData
	err = decoder.Decode(&data)
	if err != nil {
		return
	}
	list, err = data.unwrap()
	return
}

// wrap is the method used internally to convert a list of values of the
// 'plan' value to a JSON document.
func (l *PlanList) wrap() (data planListData, err error) {
	if l == nil {
		return
	}
	data = make(planListData, len(l.items))
	for i, item := range l.items {
		data[i], err = item.wrap()
		if err != nil {
			return
		}
	}
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// list of values of the 'plan' type.
func (d planListData) unwrap() (list *PlanList, err error) {
	if d == nil {
		return
	}
	items := make([]*Plan, len(d))
	for i, item := range d {
		items[i], err = item.unwrap()
		if err != nil {
			return
		}
	}
	list = new(PlanList)
	list.items = items
	return
}

// planListLinkData is type used internally to marshal and unmarshal links
// to lists of objects of type 'plan'.
type planListLinkData struct {
	Kind  *string     "json:\"kind,omitempty\""
	HREF  *string     "json:\"href,omitempty\""
	Items []*planData "json:\"items,omitempty\""
}

// wrapLink is the method used internally to convert a list of values of the
// 'plan' value to a link.
func (l *PlanList) wrapLink() (data *planListLinkData, err error) {
	if l == nil {
		return
	}
	items := make([]*planData, len(l.items))
	for i, item := range l.items {
		items[i], err = item.wrap()
		if err != nil {
			return
		}
	}
	data = new(planListLinkData)
	data.Items = items
	data.HREF = l.href
	data.Kind = new(string)
	if l.link {
		*data.Kind = PlanListLinkKind
	} else {
		*data.Kind = PlanListKind
	}
	return
}

// unwrapLink is the function used internally to convert a JSON link to a list
// of values of the 'plan' type to a list.
func (d *planListLinkData) unwrapLink() (list *PlanList, err error) {
	if d == nil {
		return
	}
	items := make([]*Plan, len(d.Items))
	for i, item := range d.Items {
		items[i], err = item.unwrap()
		if err != nil {
			return
		}
	}
	list = new(PlanList)
	list.items = items
	list.href = d.HREF
	if d.Kind != nil {
		switch *d.Kind {
		case PlanListKind:
			list.link = false
		case PlanListLinkKind:
			list.link = true
		default:
			err = fmt.Errorf(
				"expected kind '%s' or '%s' but got '%s'",
				PlanListKind,
				PlanListLinkKind,
				*d.Kind,
			)
			return
		}
	}
	return
}
