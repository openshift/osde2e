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

// MarshalSupportCaseResponse writes a value of the 'support_case_response' type to the given writer.
func MarshalSupportCaseResponse(object *SupportCaseResponse, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeSupportCaseResponse(object, stream)
	stream.Flush()
	return stream.Error
}

// writeSupportCaseResponse writes a value of the 'support_case_response' type to the given stream.
func writeSupportCaseResponse(object *SupportCaseResponse, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(SupportCaseResponseLinkKind)
	} else {
		stream.WriteString(SupportCaseResponseKind)
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
	if object.uri != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("uri")
		stream.WriteString(*object.uri)
		count++
	}
	if object.caseNumber != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("case_number")
		stream.WriteString(*object.caseNumber)
		count++
	}
	if object.clusterId != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cluster_id")
		stream.WriteString(*object.clusterId)
		count++
	}
	if object.clusterUuid != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cluster_uuid")
		stream.WriteString(*object.clusterUuid)
		count++
	}
	if object.description != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("description")
		stream.WriteString(*object.description)
		count++
	}
	if object.severity != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("severity")
		stream.WriteString(*object.severity)
		count++
	}
	if object.status != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("status")
		stream.WriteString(*object.status)
		count++
	}
	if object.subscriptionId != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("subscription_id")
		stream.WriteString(*object.subscriptionId)
		count++
	}
	if object.summary != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("summary")
		stream.WriteString(*object.summary)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalSupportCaseResponse reads a value of the 'support_case_response' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalSupportCaseResponse(source interface{}) (object *SupportCaseResponse, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readSupportCaseResponse(iterator)
	err = iterator.Error
	return
}

// readSupportCaseResponse reads a value of the 'support_case_response' type from the given iterator.
func readSupportCaseResponse(iterator *jsoniter.Iterator) *SupportCaseResponse {
	object := &SupportCaseResponse{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == SupportCaseResponseLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "uri":
			value := iterator.ReadString()
			object.uri = &value
		case "case_number":
			value := iterator.ReadString()
			object.caseNumber = &value
		case "cluster_id":
			value := iterator.ReadString()
			object.clusterId = &value
		case "cluster_uuid":
			value := iterator.ReadString()
			object.clusterUuid = &value
		case "description":
			value := iterator.ReadString()
			object.description = &value
		case "severity":
			value := iterator.ReadString()
			object.severity = &value
		case "status":
			value := iterator.ReadString()
			object.status = &value
		case "subscription_id":
			value := iterator.ReadString()
			object.subscriptionId = &value
		case "summary":
			value := iterator.ReadString()
			object.summary = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
