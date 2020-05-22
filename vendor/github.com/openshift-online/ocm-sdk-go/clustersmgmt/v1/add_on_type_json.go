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

// MarshalAddOn writes a value of the 'add_on' type to the given writer.
func MarshalAddOn(object *AddOn, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeAddOn(object, stream)
	stream.Flush()
	return stream.Error
}

// writeAddOn writes a value of the 'add_on' type to the given stream.
func writeAddOn(object *AddOn, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(AddOnLinkKind)
	} else {
		stream.WriteString(AddOnKind)
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
	if object.docsLink != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("docs_link")
		stream.WriteString(*object.docsLink)
		count++
	}
	if object.enabled != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("enabled")
		stream.WriteBool(*object.enabled)
		count++
	}
	if object.icon != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("icon")
		stream.WriteString(*object.icon)
		count++
	}
	if object.installMode != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("install_mode")
		stream.WriteString(string(*object.installMode))
		count++
	}
	if object.label != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("label")
		stream.WriteString(*object.label)
		count++
	}
	if object.name != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("name")
		stream.WriteString(*object.name)
		count++
	}
	if object.operatorName != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("operator_name")
		stream.WriteString(*object.operatorName)
		count++
	}
	if object.resourceCost != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("resource_cost")
		stream.WriteFloat64(*object.resourceCost)
		count++
	}
	if object.resourceName != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("resource_name")
		stream.WriteString(*object.resourceName)
		count++
	}
	if object.targetNamespace != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("target_namespace")
		stream.WriteString(*object.targetNamespace)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalAddOn reads a value of the 'add_on' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalAddOn(source interface{}) (object *AddOn, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readAddOn(iterator)
	err = iterator.Error
	return
}

// readAddOn reads a value of the 'add_on' type from the given iterator.
func readAddOn(iterator *jsoniter.Iterator) *AddOn {
	object := &AddOn{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == AddOnLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "description":
			value := iterator.ReadString()
			object.description = &value
		case "docs_link":
			value := iterator.ReadString()
			object.docsLink = &value
		case "enabled":
			value := iterator.ReadBool()
			object.enabled = &value
		case "icon":
			value := iterator.ReadString()
			object.icon = &value
		case "install_mode":
			text := iterator.ReadString()
			value := AddOnInstallMode(text)
			object.installMode = &value
		case "label":
			value := iterator.ReadString()
			object.label = &value
		case "name":
			value := iterator.ReadString()
			object.name = &value
		case "operator_name":
			value := iterator.ReadString()
			object.operatorName = &value
		case "resource_cost":
			value := iterator.ReadFloat64()
			object.resourceCost = &value
		case "resource_name":
			value := iterator.ReadString()
			object.resourceName = &value
		case "target_namespace":
			value := iterator.ReadString()
			object.targetNamespace = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
