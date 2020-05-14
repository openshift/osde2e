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

package v1 // github.com/openshift-online/ocm-sdk-go/authorizations/v1

import (
	"io"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalResourceReview writes a value of the 'resource_review' type to the given writer.
func MarshalResourceReview(object *ResourceReview, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeResourceReview(object, stream)
	stream.Flush()
	return stream.Error
}

// writeResourceReview writes a value of the 'resource_review' type to the given stream.
func writeResourceReview(object *ResourceReview, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if object.accountUsername != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("account_username")
		stream.WriteString(*object.accountUsername)
		count++
	}
	if object.action != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("action")
		stream.WriteString(*object.action)
		count++
	}
	if object.clusterIDs != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cluster_ids")
		writeStringList(object.clusterIDs, stream)
		count++
	}
	if object.clusterUUIDs != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cluster_uuids")
		writeStringList(object.clusterUUIDs, stream)
		count++
	}
	if object.organizationIDs != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("organization_ids")
		writeStringList(object.organizationIDs, stream)
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
	if object.subscriptionIDs != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("subscription_ids")
		writeStringList(object.subscriptionIDs, stream)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalResourceReview reads a value of the 'resource_review' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalResourceReview(source interface{}) (object *ResourceReview, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readResourceReview(iterator)
	err = iterator.Error
	return
}

// readResourceReview reads a value of the 'resource_review' type from the given iterator.
func readResourceReview(iterator *jsoniter.Iterator) *ResourceReview {
	object := &ResourceReview{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "account_username":
			value := iterator.ReadString()
			object.accountUsername = &value
		case "action":
			value := iterator.ReadString()
			object.action = &value
		case "cluster_ids":
			value := readStringList(iterator)
			object.clusterIDs = value
		case "cluster_uuids":
			value := readStringList(iterator)
			object.clusterUUIDs = value
		case "organization_ids":
			value := readStringList(iterator)
			object.organizationIDs = value
		case "resource_type":
			value := iterator.ReadString()
			object.resourceType = &value
		case "subscription_ids":
			value := readStringList(iterator)
			object.subscriptionIDs = value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
