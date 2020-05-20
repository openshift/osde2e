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

// MarshalClusterRegistration writes a value of the 'cluster_registration' type to the given writer.
func MarshalClusterRegistration(object *ClusterRegistration, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeClusterRegistration(object, stream)
	stream.Flush()
	return stream.Error
}

// writeClusterRegistration writes a value of the 'cluster_registration' type to the given stream.
func writeClusterRegistration(object *ClusterRegistration, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if object.externalID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("external_id")
		stream.WriteString(*object.externalID)
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

// UnmarshalClusterRegistration reads a value of the 'cluster_registration' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalClusterRegistration(source interface{}) (object *ClusterRegistration, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readClusterRegistration(iterator)
	err = iterator.Error
	return
}

// readClusterRegistration reads a value of the 'cluster_registration' type from the given iterator.
func readClusterRegistration(iterator *jsoniter.Iterator) *ClusterRegistration {
	object := &ClusterRegistration{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "external_id":
			value := iterator.ReadString()
			object.externalID = &value
		case "subscription_id":
			value := iterator.ReadString()
			object.subscriptionID = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
