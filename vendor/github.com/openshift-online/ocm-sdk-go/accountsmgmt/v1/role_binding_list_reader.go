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

// roleBindingListData is type used internally to marshal and unmarshal lists of objects
// of type 'role_binding'.
type roleBindingListData []*roleBindingData

// UnmarshalRoleBindingList reads a list of values of the 'role_binding'
// from the given source, which can be a slice of bytes, a string, an io.Reader or a
// json.Decoder.
func UnmarshalRoleBindingList(source interface{}) (list *RoleBindingList, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	var data roleBindingListData
	err = decoder.Decode(&data)
	if err != nil {
		return
	}
	list, err = data.unwrap()
	return
}

// wrap is the method used internally to convert a list of values of the
// 'role_binding' value to a JSON document.
func (l *RoleBindingList) wrap() (data roleBindingListData, err error) {
	if l == nil {
		return
	}
	data = make(roleBindingListData, len(l.items))
	for i, item := range l.items {
		data[i], err = item.wrap()
		if err != nil {
			return
		}
	}
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// list of values of the 'role_binding' type.
func (d roleBindingListData) unwrap() (list *RoleBindingList, err error) {
	if d == nil {
		return
	}
	items := make([]*RoleBinding, len(d))
	for i, item := range d {
		items[i], err = item.unwrap()
		if err != nil {
			return
		}
	}
	list = new(RoleBindingList)
	list.items = items
	return
}

// roleBindingListLinkData is type used internally to marshal and unmarshal links
// to lists of objects of type 'role_binding'.
type roleBindingListLinkData struct {
	Kind  *string            "json:\"kind,omitempty\""
	HREF  *string            "json:\"href,omitempty\""
	Items []*roleBindingData "json:\"items,omitempty\""
}

// wrapLink is the method used internally to convert a list of values of the
// 'role_binding' value to a link.
func (l *RoleBindingList) wrapLink() (data *roleBindingListLinkData, err error) {
	if l == nil {
		return
	}
	items := make([]*roleBindingData, len(l.items))
	for i, item := range l.items {
		items[i], err = item.wrap()
		if err != nil {
			return
		}
	}
	data = new(roleBindingListLinkData)
	data.Items = items
	data.HREF = l.href
	data.Kind = new(string)
	if l.link {
		*data.Kind = RoleBindingListLinkKind
	} else {
		*data.Kind = RoleBindingListKind
	}
	return
}

// unwrapLink is the function used internally to convert a JSON link to a list
// of values of the 'role_binding' type to a list.
func (d *roleBindingListLinkData) unwrapLink() (list *RoleBindingList, err error) {
	if d == nil {
		return
	}
	items := make([]*RoleBinding, len(d.Items))
	for i, item := range d.Items {
		items[i], err = item.unwrap()
		if err != nil {
			return
		}
	}
	list = new(RoleBindingList)
	list.items = items
	list.href = d.HREF
	if d.Kind != nil {
		list.link = *d.Kind == RoleBindingListLinkKind
	}
	return
}
