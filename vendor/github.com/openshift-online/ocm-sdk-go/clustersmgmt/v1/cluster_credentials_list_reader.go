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

// clusterCredentialsListData is type used internally to marshal and unmarshal lists of objects
// of type 'cluster_credentials'.
type clusterCredentialsListData []*clusterCredentialsData

// UnmarshalClusterCredentialsList reads a list of values of the 'cluster_credentials'
// from the given source, which can be a slice of bytes, a string, an io.Reader or a
// json.Decoder.
func UnmarshalClusterCredentialsList(source interface{}) (list *ClusterCredentialsList, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	var data clusterCredentialsListData
	err = decoder.Decode(&data)
	if err != nil {
		return
	}
	list, err = data.unwrap()
	return
}

// wrap is the method used internally to convert a list of values of the
// 'cluster_credentials' value to a JSON document.
func (l *ClusterCredentialsList) wrap() (data clusterCredentialsListData, err error) {
	if l == nil {
		return
	}
	data = make(clusterCredentialsListData, len(l.items))
	for i, item := range l.items {
		data[i], err = item.wrap()
		if err != nil {
			return
		}
	}
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// list of values of the 'cluster_credentials' type.
func (d clusterCredentialsListData) unwrap() (list *ClusterCredentialsList, err error) {
	if d == nil {
		return
	}
	items := make([]*ClusterCredentials, len(d))
	for i, item := range d {
		items[i], err = item.unwrap()
		if err != nil {
			return
		}
	}
	list = new(ClusterCredentialsList)
	list.items = items
	return
}

// clusterCredentialsListLinkData is type used internally to marshal and unmarshal links
// to lists of objects of type 'cluster_credentials'.
type clusterCredentialsListLinkData struct {
	Kind  *string                   "json:\"kind,omitempty\""
	HREF  *string                   "json:\"href,omitempty\""
	Items []*clusterCredentialsData "json:\"items,omitempty\""
}

// wrapLink is the method used internally to convert a list of values of the
// 'cluster_credentials' value to a link.
func (l *ClusterCredentialsList) wrapLink() (data *clusterCredentialsListLinkData, err error) {
	if l == nil {
		return
	}
	items := make([]*clusterCredentialsData, len(l.items))
	for i, item := range l.items {
		items[i], err = item.wrap()
		if err != nil {
			return
		}
	}
	data = new(clusterCredentialsListLinkData)
	data.Items = items
	data.HREF = l.href
	data.Kind = new(string)
	if l.link {
		*data.Kind = ClusterCredentialsListLinkKind
	} else {
		*data.Kind = ClusterCredentialsListKind
	}
	return
}

// unwrapLink is the function used internally to convert a JSON link to a list
// of values of the 'cluster_credentials' type to a list.
func (d *clusterCredentialsListLinkData) unwrapLink() (list *ClusterCredentialsList, err error) {
	if d == nil {
		return
	}
	items := make([]*ClusterCredentials, len(d.Items))
	for i, item := range d.Items {
		items[i], err = item.unwrap()
		if err != nil {
			return
		}
	}
	list = new(ClusterCredentialsList)
	list.items = items
	list.href = d.HREF
	if d.Kind != nil {
		list.link = *d.Kind == ClusterCredentialsListLinkKind
	}
	return
}
