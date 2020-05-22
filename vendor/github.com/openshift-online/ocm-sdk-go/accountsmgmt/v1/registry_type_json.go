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
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalRegistry writes a value of the 'registry' type to the given writer.
func MarshalRegistry(object *Registry, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeRegistry(object, stream)
	stream.Flush()
	return stream.Error
}

// writeRegistry writes a value of the 'registry' type to the given stream.
func writeRegistry(object *Registry, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(RegistryLinkKind)
	} else {
		stream.WriteString(RegistryKind)
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
	if object.url != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("url")
		stream.WriteString(*object.url)
		count++
	}
	if object.cloudAlias != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cloud_alias")
		stream.WriteBool(*object.cloudAlias)
		count++
	}
	if object.createdAt != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("created_at")
		stream.WriteString((*object.createdAt).Format(time.RFC3339))
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
	if object.orgName != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("org_name")
		stream.WriteString(*object.orgName)
		count++
	}
	if object.teamName != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("team_name")
		stream.WriteString(*object.teamName)
		count++
	}
	if object.type_ != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("type")
		stream.WriteString(*object.type_)
		count++
	}
	if object.updatedAt != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("updated_at")
		stream.WriteString((*object.updatedAt).Format(time.RFC3339))
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalRegistry reads a value of the 'registry' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalRegistry(source interface{}) (object *Registry, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readRegistry(iterator)
	err = iterator.Error
	return
}

// readRegistry reads a value of the 'registry' type from the given iterator.
func readRegistry(iterator *jsoniter.Iterator) *Registry {
	object := &Registry{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == RegistryLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "url":
			value := iterator.ReadString()
			object.url = &value
		case "cloud_alias":
			value := iterator.ReadBool()
			object.cloudAlias = &value
		case "created_at":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.createdAt = &value
		case "name":
			value := iterator.ReadString()
			object.name = &value
		case "org_name":
			value := iterator.ReadString()
			object.orgName = &value
		case "team_name":
			value := iterator.ReadString()
			object.teamName = &value
		case "type":
			value := iterator.ReadString()
			object.type_ = &value
		case "updated_at":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.updatedAt = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
