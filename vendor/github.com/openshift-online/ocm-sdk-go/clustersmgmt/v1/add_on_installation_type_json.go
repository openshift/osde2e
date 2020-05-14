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
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalAddOnInstallation writes a value of the 'add_on_installation' type to the given writer.
func MarshalAddOnInstallation(object *AddOnInstallation, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeAddOnInstallation(object, stream)
	stream.Flush()
	return stream.Error
}

// writeAddOnInstallation writes a value of the 'add_on_installation' type to the given stream.
func writeAddOnInstallation(object *AddOnInstallation, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(AddOnInstallationLinkKind)
	} else {
		stream.WriteString(AddOnInstallationKind)
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
	if object.addon != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("addon")
		writeAddOn(object.addon, stream)
		count++
	}
	if object.cluster != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cluster")
		writeCluster(object.cluster, stream)
		count++
	}
	if object.creationTimestamp != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("creation_timestamp")
		stream.WriteString((*object.creationTimestamp).Format(time.RFC3339))
		count++
	}
	if object.operatorVersion != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("operator_version")
		stream.WriteString(*object.operatorVersion)
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
	if object.updatedTimestamp != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("updated_timestamp")
		stream.WriteString((*object.updatedTimestamp).Format(time.RFC3339))
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalAddOnInstallation reads a value of the 'add_on_installation' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalAddOnInstallation(source interface{}) (object *AddOnInstallation, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readAddOnInstallation(iterator)
	err = iterator.Error
	return
}

// readAddOnInstallation reads a value of the 'add_on_installation' type from the given iterator.
func readAddOnInstallation(iterator *jsoniter.Iterator) *AddOnInstallation {
	object := &AddOnInstallation{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == AddOnInstallationLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "addon":
			value := readAddOn(iterator)
			object.addon = value
		case "cluster":
			value := readCluster(iterator)
			object.cluster = value
		case "creation_timestamp":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.creationTimestamp = &value
		case "operator_version":
			value := iterator.ReadString()
			object.operatorVersion = &value
		case "state":
			text := iterator.ReadString()
			value := AddOnInstallationState(text)
			object.state = &value
		case "state_description":
			value := iterator.ReadString()
			object.stateDescription = &value
		case "updated_timestamp":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.updatedTimestamp = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
