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

// MarshalResourceQuota writes a value of the 'resource_quota' type to the given writer.
func MarshalResourceQuota(object *ResourceQuota, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeResourceQuota(object, stream)
	stream.Flush()
	return stream.Error
}

// writeResourceQuota writes a value of the 'resource_quota' type to the given stream.
func writeResourceQuota(object *ResourceQuota, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(ResourceQuotaLinkKind)
	} else {
		stream.WriteString(ResourceQuotaKind)
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
	if object.byoc != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("byoc")
		stream.WriteBool(*object.byoc)
		count++
	}
	if object.sku != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("sku")
		stream.WriteString(*object.sku)
		count++
	}
	if object.allowed != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("allowed")
		stream.WriteInt(*object.allowed)
		count++
	}
	if object.availabilityZoneType != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("availability_zone_type")
		stream.WriteString(*object.availabilityZoneType)
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
	if object.organizationID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("organization_id")
		stream.WriteString(*object.organizationID)
		count++
	}
	if object.resourceName != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("resource_name")
		stream.WriteString(*object.resourceName)
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
	if object.skuCount != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("sku_count")
		stream.WriteInt(*object.skuCount)
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

// UnmarshalResourceQuota reads a value of the 'resource_quota' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalResourceQuota(source interface{}) (object *ResourceQuota, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readResourceQuota(iterator)
	err = iterator.Error
	return
}

// readResourceQuota reads a value of the 'resource_quota' type from the given iterator.
func readResourceQuota(iterator *jsoniter.Iterator) *ResourceQuota {
	object := &ResourceQuota{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == ResourceQuotaLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "byoc":
			value := iterator.ReadBool()
			object.byoc = &value
		case "sku":
			value := iterator.ReadString()
			object.sku = &value
		case "allowed":
			value := iterator.ReadInt()
			object.allowed = &value
		case "availability_zone_type":
			value := iterator.ReadString()
			object.availabilityZoneType = &value
		case "created_at":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.createdAt = &value
		case "organization_id":
			value := iterator.ReadString()
			object.organizationID = &value
		case "resource_name":
			value := iterator.ReadString()
			object.resourceName = &value
		case "resource_type":
			value := iterator.ReadString()
			object.resourceType = &value
		case "sku_count":
			value := iterator.ReadInt()
			object.skuCount = &value
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
