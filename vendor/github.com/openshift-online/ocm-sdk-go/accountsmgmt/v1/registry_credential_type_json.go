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

// MarshalRegistryCredential writes a value of the 'registry_credential' type to the given writer.
func MarshalRegistryCredential(object *RegistryCredential, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeRegistryCredential(object, stream)
	stream.Flush()
	return stream.Error
}

// writeRegistryCredential writes a value of the 'registry_credential' type to the given stream.
func writeRegistryCredential(object *RegistryCredential, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(RegistryCredentialLinkKind)
	} else {
		stream.WriteString(RegistryCredentialKind)
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
	if object.account != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("account")
		writeAccount(object.account, stream)
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
	if object.externalResourceID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("external_resource_id")
		stream.WriteString(*object.externalResourceID)
		count++
	}
	if object.registry != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("registry")
		writeRegistry(object.registry, stream)
		count++
	}
	if object.token != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("token")
		stream.WriteString(*object.token)
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
	if object.username != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("username")
		stream.WriteString(*object.username)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalRegistryCredential reads a value of the 'registry_credential' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalRegistryCredential(source interface{}) (object *RegistryCredential, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readRegistryCredential(iterator)
	err = iterator.Error
	return
}

// readRegistryCredential reads a value of the 'registry_credential' type from the given iterator.
func readRegistryCredential(iterator *jsoniter.Iterator) *RegistryCredential {
	object := &RegistryCredential{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == RegistryCredentialLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "account":
			value := readAccount(iterator)
			object.account = value
		case "created_at":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.createdAt = &value
		case "external_resource_id":
			value := iterator.ReadString()
			object.externalResourceID = &value
		case "registry":
			value := readRegistry(iterator)
			object.registry = value
		case "token":
			value := iterator.ReadString()
			object.token = &value
		case "updated_at":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.updatedAt = &value
		case "username":
			value := iterator.ReadString()
			object.username = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
