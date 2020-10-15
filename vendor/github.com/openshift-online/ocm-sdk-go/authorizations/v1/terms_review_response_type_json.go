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

// MarshalTermsReviewResponse writes a value of the 'terms_review_response' type to the given writer.
func MarshalTermsReviewResponse(object *TermsReviewResponse, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeTermsReviewResponse(object, stream)
	stream.Flush()
	return stream.Error
}

// writeTermsReviewResponse writes a value of the 'terms_review_response' type to the given stream.
func writeTermsReviewResponse(object *TermsReviewResponse, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if object.accountId != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("account_id")
		stream.WriteString(*object.accountId)
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
	if object.redirectUrl != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("redirect_url")
		stream.WriteString(*object.redirectUrl)
		count++
	}
	if object.termsAvailable != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("terms_available")
		stream.WriteBool(*object.termsAvailable)
		count++
	}
	if object.termsRequired != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("terms_required")
		stream.WriteBool(*object.termsRequired)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalTermsReviewResponse reads a value of the 'terms_review_response' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalTermsReviewResponse(source interface{}) (object *TermsReviewResponse, err error) {
	if source == http.NoBody {
		return
	}
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readTermsReviewResponse(iterator)
	err = iterator.Error
	return
}

// readTermsReviewResponse reads a value of the 'terms_review_response' type from the given iterator.
func readTermsReviewResponse(iterator *jsoniter.Iterator) *TermsReviewResponse {
	object := &TermsReviewResponse{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "account_id":
			value := iterator.ReadString()
			object.accountId = &value
		case "organization_id":
			value := iterator.ReadString()
			object.organizationID = &value
		case "redirect_url":
			value := iterator.ReadString()
			object.redirectUrl = &value
		case "terms_available":
			value := iterator.ReadBool()
			object.termsAvailable = &value
		case "terms_required":
			value := iterator.ReadBool()
			object.termsRequired = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
