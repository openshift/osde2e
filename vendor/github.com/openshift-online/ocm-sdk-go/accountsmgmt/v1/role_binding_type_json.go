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

// MarshalRoleBinding writes a value of the 'role_binding' type to the given writer.
func MarshalRoleBinding(object *RoleBinding, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeRoleBinding(object, stream)
	stream.Flush()
	return stream.Error
}

// writeRoleBinding writes a value of the 'role_binding' type to the given stream.
func writeRoleBinding(object *RoleBinding, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(RoleBindingLinkKind)
	} else {
		stream.WriteString(RoleBindingKind)
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
	if object.accountID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("account_id")
		stream.WriteString(*object.accountID)
		count++
	}
	if object.configManaged != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("config_managed")
		stream.WriteBool(*object.configManaged)
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
	if object.organization != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("organization")
		writeOrganization(object.organization, stream)
		count++
	}
	if object.organizationID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("organization_id")
		stream.WriteString(*object.organizationID)
		count++
	}
	if object.role != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("role")
		writeRole(object.role, stream)
		count++
	}
	if object.roleID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("role_id")
		stream.WriteString(*object.roleID)
		count++
	}
	if object.subscription != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("subscription")
		writeSubscription(object.subscription, stream)
		count++
	}
	if object.subscriptionID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("subscription_id")
		stream.WriteString(*object.subscriptionID)
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

// UnmarshalRoleBinding reads a value of the 'role_binding' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalRoleBinding(source interface{}) (object *RoleBinding, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readRoleBinding(iterator)
	err = iterator.Error
	return
}

// readRoleBinding reads a value of the 'role_binding' type from the given iterator.
func readRoleBinding(iterator *jsoniter.Iterator) *RoleBinding {
	object := &RoleBinding{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == RoleBindingLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "account":
			value := readAccount(iterator)
			object.account = value
		case "account_id":
			value := iterator.ReadString()
			object.accountID = &value
		case "config_managed":
			value := iterator.ReadBool()
			object.configManaged = &value
		case "created_at":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.createdAt = &value
		case "organization":
			value := readOrganization(iterator)
			object.organization = value
		case "organization_id":
			value := iterator.ReadString()
			object.organizationID = &value
		case "role":
			value := readRole(iterator)
			object.role = value
		case "role_id":
			value := iterator.ReadString()
			object.roleID = &value
		case "subscription":
			value := readSubscription(iterator)
			object.subscription = value
		case "subscription_id":
			value := iterator.ReadString()
			object.subscriptionID = &value
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
