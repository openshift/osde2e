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

package v1 // github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1

import (
	"io"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalAWSInfrastructureAccessRole writes a value of the 'AWS_infrastructure_access_role' type to the given writer.
func MarshalAWSInfrastructureAccessRole(object *AWSInfrastructureAccessRole, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeAWSInfrastructureAccessRole(object, stream)
	stream.Flush()
	return stream.Error
}

// writeAWSInfrastructureAccessRole writes a value of the 'AWS_infrastructure_access_role' type to the given stream.
func writeAWSInfrastructureAccessRole(object *AWSInfrastructureAccessRole, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(AWSInfrastructureAccessRoleLinkKind)
	} else {
		stream.WriteString(AWSInfrastructureAccessRoleKind)
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
	if object.description != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("description")
		stream.WriteString(*object.description)
		count++
	}
	if object.displayName != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("display_name")
		stream.WriteString(*object.displayName)
		count++
	}
	if object.state != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("state")
		stream.WriteString(string(*object.state))
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalAWSInfrastructureAccessRole reads a value of the 'AWS_infrastructure_access_role' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalAWSInfrastructureAccessRole(source interface{}) (object *AWSInfrastructureAccessRole, err error) {
	if source == http.NoBody {
		return
	}
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readAWSInfrastructureAccessRole(iterator)
	err = iterator.Error
	return
}

// readAWSInfrastructureAccessRole reads a value of the 'AWS_infrastructure_access_role' type from the given iterator.
func readAWSInfrastructureAccessRole(iterator *jsoniter.Iterator) *AWSInfrastructureAccessRole {
	object := &AWSInfrastructureAccessRole{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == AWSInfrastructureAccessRoleLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "description":
			value := iterator.ReadString()
			object.description = &value
		case "display_name":
			value := iterator.ReadString()
			object.displayName = &value
		case "state":
			text := iterator.ReadString()
			value := AWSInfrastructureAccessRoleState(text)
			object.state = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
