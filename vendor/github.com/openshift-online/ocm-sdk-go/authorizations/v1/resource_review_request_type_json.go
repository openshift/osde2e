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

// MarshalResourceReviewRequest writes a value of the 'resource_review_request' type to the given writer.
func MarshalResourceReviewRequest(object *ResourceReviewRequest, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeResourceReviewRequest(object, stream)
	stream.Flush()
	return stream.Error
}

// writeResourceReviewRequest writes a value of the 'resource_review_request' type to the given stream.
func writeResourceReviewRequest(object *ResourceReviewRequest, stream *jsoniter.Stream) {
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

// UnmarshalResourceReviewRequest reads a value of the 'resource_review_request' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalResourceReviewRequest(source interface{}) (object *ResourceReviewRequest, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readResourceReviewRequest(iterator)
	err = iterator.Error
	return
}

// readResourceReviewRequest reads a value of the 'resource_review_request' type from the given iterator.
func readResourceReviewRequest(iterator *jsoniter.Iterator) *ResourceReviewRequest {
	object := &ResourceReviewRequest{}
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
		case "resource_type":
			value := iterator.ReadString()
			object.resourceType = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
