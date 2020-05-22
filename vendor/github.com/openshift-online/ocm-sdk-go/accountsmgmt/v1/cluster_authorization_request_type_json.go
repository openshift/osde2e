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

// MarshalClusterAuthorizationRequest writes a value of the 'cluster_authorization_request' type to the given writer.
func MarshalClusterAuthorizationRequest(object *ClusterAuthorizationRequest, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeClusterAuthorizationRequest(object, stream)
	stream.Flush()
	return stream.Error
}

// writeClusterAuthorizationRequest writes a value of the 'cluster_authorization_request' type to the given stream.
func writeClusterAuthorizationRequest(object *ClusterAuthorizationRequest, stream *jsoniter.Stream) {
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
	if object.accountUsername != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("account_username")
		stream.WriteString(*object.accountUsername)
		count++
	}
	if object.availabilityZone != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("availability_zone")
		stream.WriteString(*object.availabilityZone)
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
	if object.disconnected != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("disconnected")
		stream.WriteBool(*object.disconnected)
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
	if object.externalClusterID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("external_cluster_id")
		stream.WriteString(*object.externalClusterID)
		count++
	}
	if object.managed != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("managed")
		stream.WriteBool(*object.managed)
		count++
	}
	if object.reserve != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("reserve")
		stream.WriteBool(*object.reserve)
		count++
	}
	if object.resources != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("resources")
		writeReservedResourceList(object.resources, stream)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalClusterAuthorizationRequest reads a value of the 'cluster_authorization_request' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalClusterAuthorizationRequest(source interface{}) (object *ClusterAuthorizationRequest, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readClusterAuthorizationRequest(iterator)
	err = iterator.Error
	return
}

// readClusterAuthorizationRequest reads a value of the 'cluster_authorization_request' type from the given iterator.
func readClusterAuthorizationRequest(iterator *jsoniter.Iterator) *ClusterAuthorizationRequest {
	object := &ClusterAuthorizationRequest{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "byoc":
			value := iterator.ReadBool()
			object.byoc = &value
		case "account_username":
			value := iterator.ReadString()
			object.accountUsername = &value
		case "availability_zone":
			value := iterator.ReadString()
			object.availabilityZone = &value
		case "cluster_id":
			value := iterator.ReadString()
			object.clusterID = &value
		case "disconnected":
			value := iterator.ReadBool()
			object.disconnected = &value
		case "display_name":
			value := iterator.ReadString()
			object.displayName = &value
		case "external_cluster_id":
			value := iterator.ReadString()
			object.externalClusterID = &value
		case "managed":
			value := iterator.ReadBool()
			object.managed = &value
		case "reserve":
			value := iterator.ReadBool()
			object.reserve = &value
		case "resources":
			value := readReservedResourceList(iterator)
			object.resources = value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
