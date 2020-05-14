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

// MarshalQuotaSummary writes a value of the 'quota_summary' type to the given writer.
func MarshalQuotaSummary(object *QuotaSummary, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeQuotaSummary(object, stream)
	stream.Flush()
	return stream.Error
}

// writeQuotaSummary writes a value of the 'quota_summary' type to the given stream.
func writeQuotaSummary(object *QuotaSummary, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if object.byoc != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("byoc")
		stream.WriteBool(*object.byoc)
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
	if object.organizationID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("organization_id")
		stream.WriteString(*object.organizationID)
		count++
	}
	if object.reserved != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("reserved")
		stream.WriteInt(*object.reserved)
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
	stream.WriteObjectEnd()
}

// UnmarshalQuotaSummary reads a value of the 'quota_summary' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalQuotaSummary(source interface{}) (object *QuotaSummary, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readQuotaSummary(iterator)
	err = iterator.Error
	return
}

// readQuotaSummary reads a value of the 'quota_summary' type from the given iterator.
func readQuotaSummary(iterator *jsoniter.Iterator) *QuotaSummary {
	object := &QuotaSummary{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "byoc":
			value := iterator.ReadBool()
			object.byoc = &value
		case "allowed":
			value := iterator.ReadInt()
			object.allowed = &value
		case "availability_zone_type":
			value := iterator.ReadString()
			object.availabilityZoneType = &value
		case "organization_id":
			value := iterator.ReadString()
			object.organizationID = &value
		case "reserved":
			value := iterator.ReadInt()
			object.reserved = &value
		case "resource_name":
			value := iterator.ReadString()
			object.resourceName = &value
		case "resource_type":
			value := iterator.ReadString()
			object.resourceType = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
