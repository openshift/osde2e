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

// MarshalAccount writes a value of the 'account' type to the given writer.
func MarshalAccount(object *Account, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeAccount(object, stream)
	stream.Flush()
	return stream.Error
}

// writeAccount writes a value of the 'account' type to the given stream.
func writeAccount(object *Account, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(AccountLinkKind)
	} else {
		stream.WriteString(AccountKind)
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
	if object.banCode != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("ban_code")
		stream.WriteString(*object.banCode)
		count++
	}
	if object.banDescription != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("ban_description")
		stream.WriteString(*object.banDescription)
		count++
	}
	if object.banned != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("banned")
		stream.WriteBool(*object.banned)
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
	if object.email != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("email")
		stream.WriteString(*object.email)
		count++
	}
	if object.firstName != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("first_name")
		stream.WriteString(*object.firstName)
		count++
	}
	if object.labels != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("labels")
		writeLabelList(object.labels, stream)
		count++
	}
	if object.lastName != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("last_name")
		stream.WriteString(*object.lastName)
		count++
	}
	if object.organization != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("organization")
		writeOrganization(object.organization, stream)
		count++
	}
	if object.serviceAccount != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("service_account")
		stream.WriteBool(*object.serviceAccount)
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

// UnmarshalAccount reads a value of the 'account' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalAccount(source interface{}) (object *Account, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readAccount(iterator)
	err = iterator.Error
	return
}

// readAccount reads a value of the 'account' type from the given iterator.
func readAccount(iterator *jsoniter.Iterator) *Account {
	object := &Account{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == AccountLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "ban_code":
			value := iterator.ReadString()
			object.banCode = &value
		case "ban_description":
			value := iterator.ReadString()
			object.banDescription = &value
		case "banned":
			value := iterator.ReadBool()
			object.banned = &value
		case "created_at":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.createdAt = &value
		case "email":
			value := iterator.ReadString()
			object.email = &value
		case "first_name":
			value := iterator.ReadString()
			object.firstName = &value
		case "labels":
			value := readLabelList(iterator)
			object.labels = value
		case "last_name":
			value := iterator.ReadString()
			object.lastName = &value
		case "organization":
			value := readOrganization(iterator)
			object.organization = value
		case "service_account":
			value := iterator.ReadBool()
			object.serviceAccount = &value
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
