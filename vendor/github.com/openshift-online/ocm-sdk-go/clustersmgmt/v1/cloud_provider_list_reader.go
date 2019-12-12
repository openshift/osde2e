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
	"fmt"

	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// cloudProviderListData is type used internally to marshal and unmarshal lists of objects
// of type 'cloud_provider'.
type cloudProviderListData []*cloudProviderData

// UnmarshalCloudProviderList reads a list of values of the 'cloud_provider'
// from the given source, which can be a slice of bytes, a string, an io.Reader or a
// json.Decoder.
func UnmarshalCloudProviderList(source interface{}) (list *CloudProviderList, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	var data cloudProviderListData
	err = decoder.Decode(&data)
	if err != nil {
		return
	}
	list, err = data.unwrap()
	return
}

// wrap is the method used internally to convert a list of values of the
// 'cloud_provider' value to a JSON document.
func (l *CloudProviderList) wrap() (data cloudProviderListData, err error) {
	if l == nil {
		return
	}
	data = make(cloudProviderListData, len(l.items))
	for i, item := range l.items {
		data[i], err = item.wrap()
		if err != nil {
			return
		}
	}
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// list of values of the 'cloud_provider' type.
func (d cloudProviderListData) unwrap() (list *CloudProviderList, err error) {
	if d == nil {
		return
	}
	items := make([]*CloudProvider, len(d))
	for i, item := range d {
		items[i], err = item.unwrap()
		if err != nil {
			return
		}
	}
	list = new(CloudProviderList)
	list.items = items
	return
}

// cloudProviderListLinkData is type used internally to marshal and unmarshal links
// to lists of objects of type 'cloud_provider'.
type cloudProviderListLinkData struct {
	Kind  *string              "json:\"kind,omitempty\""
	HREF  *string              "json:\"href,omitempty\""
	Items []*cloudProviderData "json:\"items,omitempty\""
}

// wrapLink is the method used internally to convert a list of values of the
// 'cloud_provider' value to a link.
func (l *CloudProviderList) wrapLink() (data *cloudProviderListLinkData, err error) {
	if l == nil {
		return
	}
	items := make([]*cloudProviderData, len(l.items))
	for i, item := range l.items {
		items[i], err = item.wrap()
		if err != nil {
			return
		}
	}
	data = new(cloudProviderListLinkData)
	data.Items = items
	data.HREF = l.href
	data.Kind = new(string)
	if l.link {
		*data.Kind = CloudProviderListLinkKind
	} else {
		*data.Kind = CloudProviderListKind
	}
	return
}

// unwrapLink is the function used internally to convert a JSON link to a list
// of values of the 'cloud_provider' type to a list.
func (d *cloudProviderListLinkData) unwrapLink() (list *CloudProviderList, err error) {
	if d == nil {
		return
	}
	items := make([]*CloudProvider, len(d.Items))
	for i, item := range d.Items {
		items[i], err = item.unwrap()
		if err != nil {
			return
		}
	}
	list = new(CloudProviderList)
	list.items = items
	list.href = d.HREF
	if d.Kind != nil {
		switch *d.Kind {
		case CloudProviderListKind:
			list.link = false
		case CloudProviderListLinkKind:
			list.link = true
		default:
			err = fmt.Errorf(
				"expected kind '%s' or '%s' but got '%s'",
				CloudProviderListKind,
				CloudProviderListLinkKind,
				*d.Kind,
			)
			return
		}
	}
	return
}
