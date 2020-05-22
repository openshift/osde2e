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

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalAWSInfrastructureAccessRoleGrant writes a value of the 'AWS_infrastructure_access_role_grant' type to the given writer.
func MarshalAWSInfrastructureAccessRoleGrant(object *AWSInfrastructureAccessRoleGrant, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeAWSInfrastructureAccessRoleGrant(object, stream)
	stream.Flush()
	return stream.Error
}

// writeAWSInfrastructureAccessRoleGrant writes a value of the 'AWS_infrastructure_access_role_grant' type to the given stream.
func writeAWSInfrastructureAccessRoleGrant(object *AWSInfrastructureAccessRoleGrant, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(AWSInfrastructureAccessRoleGrantLinkKind)
	} else {
		stream.WriteString(AWSInfrastructureAccessRoleGrantKind)
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
	if object.consoleURL != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("console_url")
		stream.WriteString(*object.consoleURL)
		count++
	}
	if object.role != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("role")
		writeAWSInfrastructureAccessRole(object.role, stream)
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
	if object.stateDescription != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("state_description")
		stream.WriteString(*object.stateDescription)
		count++
	}
	if object.userARN != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("user_arn")
		stream.WriteString(*object.userARN)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalAWSInfrastructureAccessRoleGrant reads a value of the 'AWS_infrastructure_access_role_grant' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalAWSInfrastructureAccessRoleGrant(source interface{}) (object *AWSInfrastructureAccessRoleGrant, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readAWSInfrastructureAccessRoleGrant(iterator)
	err = iterator.Error
	return
}

// readAWSInfrastructureAccessRoleGrant reads a value of the 'AWS_infrastructure_access_role_grant' type from the given iterator.
func readAWSInfrastructureAccessRoleGrant(iterator *jsoniter.Iterator) *AWSInfrastructureAccessRoleGrant {
	object := &AWSInfrastructureAccessRoleGrant{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == AWSInfrastructureAccessRoleGrantLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "console_url":
			value := iterator.ReadString()
			object.consoleURL = &value
		case "role":
			value := readAWSInfrastructureAccessRole(iterator)
			object.role = value
		case "state":
			text := iterator.ReadString()
			value := AWSInfrastructureAccessRoleGrantState(text)
			object.state = &value
		case "state_description":
			value := iterator.ReadString()
			object.stateDescription = &value
		case "user_arn":
			value := iterator.ReadString()
			object.userARN = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
