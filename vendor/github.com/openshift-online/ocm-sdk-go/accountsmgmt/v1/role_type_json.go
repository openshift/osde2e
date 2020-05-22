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

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalRole writes a value of the 'role' type to the given writer.
func MarshalRole(object *Role, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeRole(object, stream)
	stream.Flush()
	return stream.Error
}

// writeRole writes a value of the 'role' type to the given stream.
func writeRole(object *Role, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(RoleLinkKind)
	} else {
		stream.WriteString(RoleKind)
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
	if object.name != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("name")
		stream.WriteString(*object.name)
		count++
	}
	if object.permissions != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("permissions")
		writePermissionList(object.permissions, stream)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalRole reads a value of the 'role' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalRole(source interface{}) (object *Role, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readRole(iterator)
	err = iterator.Error
	return
}

// readRole reads a value of the 'role' type from the given iterator.
func readRole(iterator *jsoniter.Iterator) *Role {
	object := &Role{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == RoleLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "name":
			value := iterator.ReadString()
			object.name = &value
		case "permissions":
			value := readPermissionList(iterator)
			object.permissions = value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
