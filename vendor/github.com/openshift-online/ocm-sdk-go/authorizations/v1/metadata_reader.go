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

// metadataData is the data structure used internally to marshal and unmarshal
// metadata.
type metadataData struct {
	ServerVersion *string "json:\"server_version,omitempty\""
}

// MarshalMetadata writes a value of the metadata type to the given target, which
// can be a writer or a JSON encoder.
func MarshalMetadata(object *Metadata, target interface{}) error {
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

// wrap is the method used internally to convert a metadata object to a JSON
// document.
func (m *Metadata) wrap() (data *metadataData, err error) {
	if m == nil {
		return
	}
	data = &metadataData{
		ServerVersion: m.serverVersion,
	}
	return
}

// UnmarshalMetadata reads a value of the metadata type from the given source, which
// which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalMetadata(source interface{}) (object *Metadata, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := &metadataData{}
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the metadata type.
func (d *metadataData) unwrap() (object *Metadata, err error) {
	if d == nil {
		return
	}
	object = &Metadata{
		serverVersion: d.ServerVersion,
	}
	return
}
