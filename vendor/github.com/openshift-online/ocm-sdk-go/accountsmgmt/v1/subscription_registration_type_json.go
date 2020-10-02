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

// MarshalSubscriptionRegistration writes a value of the 'subscription_registration' type to the given writer.
func MarshalSubscriptionRegistration(object *SubscriptionRegistration, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeSubscriptionRegistration(object, stream)
	stream.Flush()
	return stream.Error
}

// writeSubscriptionRegistration writes a value of the 'subscription_registration' type to the given stream.
func writeSubscriptionRegistration(object *SubscriptionRegistration, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if object.clusterUUID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cluster_uuid")
		stream.WriteString(*object.clusterUUID)
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
	if object.displayName != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("display_name")
		stream.WriteString(*object.displayName)
		count++
	}
	if object.planID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("plan_id")
		stream.WriteString(string(*object.planID))
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
	stream.WriteObjectEnd()
}

// UnmarshalSubscriptionRegistration reads a value of the 'subscription_registration' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalSubscriptionRegistration(source interface{}) (object *SubscriptionRegistration, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readSubscriptionRegistration(iterator)
	err = iterator.Error
	return
}

// readSubscriptionRegistration reads a value of the 'subscription_registration' type from the given iterator.
func readSubscriptionRegistration(iterator *jsoniter.Iterator) *SubscriptionRegistration {
	object := &SubscriptionRegistration{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "cluster_uuid":
			value := iterator.ReadString()
			object.clusterUUID = &value
		case "console_url":
			value := iterator.ReadString()
			object.consoleURL = &value
		case "display_name":
			value := iterator.ReadString()
			object.displayName = &value
		case "plan_id":
			text := iterator.ReadString()
			value := PlanID(text)
			object.planID = &value
		case "status":
			value := iterator.ReadString()
			object.status = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
