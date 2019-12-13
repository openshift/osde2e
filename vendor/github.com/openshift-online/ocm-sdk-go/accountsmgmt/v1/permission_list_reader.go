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

// permissionListData is type used internally to marshal and unmarshal lists of objects
// of type 'permission'.
type permissionListData []*permissionData

// UnmarshalPermissionList reads a list of values of the 'permission'
// from the given source, which can be a slice of bytes, a string, an io.Reader or a
// json.Decoder.
func UnmarshalPermissionList(source interface{}) (list *PermissionList, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	var data permissionListData
	err = decoder.Decode(&data)
	if err != nil {
		return
	}
	list, err = data.unwrap()
	return
}

// wrap is the method used internally to convert a list of values of the
// 'permission' value to a JSON document.
func (l *PermissionList) wrap() (data permissionListData, err error) {
	if l == nil {
		return
	}
	data = make(permissionListData, len(l.items))
	for i, item := range l.items {
		data[i], err = item.wrap()
		if err != nil {
			return
		}
	}
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// list of values of the 'permission' type.
func (d permissionListData) unwrap() (list *PermissionList, err error) {
	if d == nil {
		return
	}
	items := make([]*Permission, len(d))
	for i, item := range d {
		items[i], err = item.unwrap()
		if err != nil {
			return
		}
	}
	list = new(PermissionList)
	list.items = items
	return
}

// permissionListLinkData is type used internally to marshal and unmarshal links
// to lists of objects of type 'permission'.
type permissionListLinkData struct {
	Kind  *string           "json:\"kind,omitempty\""
	HREF  *string           "json:\"href,omitempty\""
	Items []*permissionData "json:\"items,omitempty\""
}

// wrapLink is the method used internally to convert a list of values of the
// 'permission' value to a link.
func (l *PermissionList) wrapLink() (data *permissionListLinkData, err error) {
	if l == nil {
		return
	}
	items := make([]*permissionData, len(l.items))
	for i, item := range l.items {
		items[i], err = item.wrap()
		if err != nil {
			return
		}
	}
	data = new(permissionListLinkData)
	data.Items = items
	data.HREF = l.href
	data.Kind = new(string)
	if l.link {
		*data.Kind = PermissionListLinkKind
	} else {
		*data.Kind = PermissionListKind
	}
	return
}

// unwrapLink is the function used internally to convert a JSON link to a list
// of values of the 'permission' type to a list.
func (d *permissionListLinkData) unwrapLink() (list *PermissionList, err error) {
	if d == nil {
		return
	}
	items := make([]*Permission, len(d.Items))
	for i, item := range d.Items {
		items[i], err = item.unwrap()
		if err != nil {
			return
		}
	}
	list = new(PermissionList)
	list.items = items
	list.href = d.HREF
	if d.Kind != nil {
		list.link = *d.Kind == PermissionListLinkKind
	}
	return
}
