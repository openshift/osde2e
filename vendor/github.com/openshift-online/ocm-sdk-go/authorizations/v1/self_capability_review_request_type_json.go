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

// MarshalSelfCapabilityReviewRequest writes a value of the 'self_capability_review_request' type to the given writer.
func MarshalSelfCapabilityReviewRequest(object *SelfCapabilityReviewRequest, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeSelfCapabilityReviewRequest(object, stream)
	stream.Flush()
	return stream.Error
}

// writeSelfCapabilityReviewRequest writes a value of the 'self_capability_review_request' type to the given stream.
func writeSelfCapabilityReviewRequest(object *SelfCapabilityReviewRequest, stream *jsoniter.Stream) {
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
	if object.capability != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("capability")
		stream.WriteString(*object.capability)
		count++
	}
	if object.clusterID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cluster_id")
		stream.WriteString(*object.clusterID)
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
	if object.resourceType != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("resource_type")
		stream.WriteString(*object.resourceType)
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
	stream.WriteObjectEnd()
}

// UnmarshalSelfCapabilityReviewRequest reads a value of the 'self_capability_review_request' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalSelfCapabilityReviewRequest(source interface{}) (object *SelfCapabilityReviewRequest, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readSelfCapabilityReviewRequest(iterator)
	err = iterator.Error
	return
}

// readSelfCapabilityReviewRequest reads a value of the 'self_capability_review_request' type from the given iterator.
func readSelfCapabilityReviewRequest(iterator *jsoniter.Iterator) *SelfCapabilityReviewRequest {
	object := &SelfCapabilityReviewRequest{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "account_username":
			value := iterator.ReadString()
			object.accountUsername = &value
		case "capability":
			value := iterator.ReadString()
			object.capability = &value
		case "cluster_id":
			value := iterator.ReadString()
			object.clusterID = &value
		case "organization_id":
			value := iterator.ReadString()
			object.organizationID = &value
		case "resource_type":
			value := iterator.ReadString()
			object.resourceType = &value
		case "subscription_id":
			value := iterator.ReadString()
			object.subscriptionID = &value
		case "type":
			value := iterator.ReadString()
			object.type_ = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
