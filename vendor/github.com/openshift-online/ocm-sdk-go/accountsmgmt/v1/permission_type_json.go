/*
Copyright (c) 2020 Red Hat, Inc.

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
	"io"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalPermission writes a value of the 'permission' type to the given writer.
func MarshalPermission(object *Permission, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writePermission(object, stream)
	stream.Flush()
	return stream.Error
}

// writePermission writes a value of the 'permission' type to the given stream.
func writePermission(object *Permission, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(PermissionLinkKind)
	} else {
		stream.WriteString(PermissionKind)
	}
	count++
	if object.id != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("id")
		stream.WriteString(*object.id)
		count++
	}
	if object.href != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("href")
		stream.WriteString(*object.href)
		count++
	}
	if object.action != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("action")
		stream.WriteString(string(*object.action))
		count++
	}
	if object.resourceType != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("resource_type")
		stream.WriteString(*object.resourceType)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalPermission reads a value of the 'permission' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalPermission(source interface{}) (object *Permission, err error) {
	if source == http.NoBody {
		return
	}
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readPermission(iterator)
	err = iterator.Error
	return
}

// readPermission reads a value of the 'permission' type from the given iterator.
func readPermission(iterator *jsoniter.Iterator) *Permission {
	object := &Permission{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == PermissionLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "action":
			text := iterator.ReadString()
			value := Action(text)
			object.action = &value
		case "resource_type":
			value := iterator.ReadString()
			object.resourceType = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
