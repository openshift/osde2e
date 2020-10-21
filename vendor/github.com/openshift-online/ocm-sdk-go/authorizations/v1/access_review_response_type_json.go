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
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalAccessReviewResponse writes a value of the 'access_review_response' type to the given writer.
func MarshalAccessReviewResponse(object *AccessReviewResponse, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeAccessReviewResponse(object, stream)
	stream.Flush()
	return stream.Error
}

// writeAccessReviewResponse writes a value of the 'access_review_response' type to the given stream.
func writeAccessReviewResponse(object *AccessReviewResponse, stream *jsoniter.Stream) {
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
	if object.allowed != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("allowed")
		stream.WriteBool(*object.allowed)
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
	if object.clusterUUID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cluster_uuid")
		stream.WriteString(*object.clusterUUID)
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
	stream.WriteObjectEnd()
}

// UnmarshalAccessReviewResponse reads a value of the 'access_review_response' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalAccessReviewResponse(source interface{}) (object *AccessReviewResponse, err error) {
	if source == http.NoBody {
		return
	}
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readAccessReviewResponse(iterator)
	err = iterator.Error
	return
}

// readAccessReviewResponse reads a value of the 'access_review_response' type from the given iterator.
func readAccessReviewResponse(iterator *jsoniter.Iterator) *AccessReviewResponse {
	object := &AccessReviewResponse{}
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
		case "allowed":
			value := iterator.ReadBool()
			object.allowed = &value
		case "cluster_id":
			value := iterator.ReadString()
			object.clusterID = &value
		case "cluster_uuid":
			value := iterator.ReadString()
			object.clusterUUID = &value
		case "organization_id":
			value := iterator.ReadString()
			object.organizationID = &value
		case "resource_type":
			value := iterator.ReadString()
			object.resourceType = &value
		case "subscription_id":
			value := iterator.ReadString()
			object.subscriptionID = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
