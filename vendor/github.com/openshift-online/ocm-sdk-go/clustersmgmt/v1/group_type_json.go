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

package v1 // github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1

import (
	"io"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalGroup writes a value of the 'group' type to the given writer.
func MarshalGroup(object *Group, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeGroup(object, stream)
	stream.Flush()
	return stream.Error
}

// writeGroup writes a value of the 'group' type to the given stream.
func writeGroup(object *Group, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(GroupLinkKind)
	} else {
		stream.WriteString(GroupKind)
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
	if object.users != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("users")
		stream.WriteObjectStart()
		stream.WriteObjectField("items")
		writeUserList(object.users.items, stream)
		stream.WriteObjectEnd()
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalGroup reads a value of the 'group' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalGroup(source interface{}) (object *Group, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readGroup(iterator)
	err = iterator.Error
	return
}

// readGroup reads a value of the 'group' type from the given iterator.
func readGroup(iterator *jsoniter.Iterator) *Group {
	object := &Group{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == GroupLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "users":
			value := &UserList{}
			for {
				field := iterator.ReadObject()
				if field == "" {
					break
				}
				switch field {
				case "kind":
					text := iterator.ReadString()
					value.link = text == UserListLinkKind
				case "href":
					text := iterator.ReadString()
					value.href = &text
				case "items":
					value.items = readUserList(iterator)
				default:
					iterator.ReadAny()
				}
			}
			object.users = value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
