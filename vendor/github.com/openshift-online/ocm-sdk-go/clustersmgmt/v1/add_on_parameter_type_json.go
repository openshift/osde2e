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

// MarshalAddOnParameter writes a value of the 'add_on_parameter' type to the given writer.
func MarshalAddOnParameter(object *AddOnParameter, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeAddOnParameter(object, stream)
	stream.Flush()
	return stream.Error
}

// writeAddOnParameter writes a value of the 'add_on_parameter' type to the given stream.
func writeAddOnParameter(object *AddOnParameter, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(AddOnParameterLinkKind)
	} else {
		stream.WriteString(AddOnParameterKind)
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
	if object.defaultValue != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("default_value")
		stream.WriteString(*object.defaultValue)
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
	if object.editable != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("editable")
		stream.WriteBool(*object.editable)
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
	if object.name != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("name")
		stream.WriteString(*object.name)
		count++
	}
	if object.required != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("required")
		stream.WriteBool(*object.required)
		count++
	}
	if object.validation != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("validation")
		stream.WriteString(*object.validation)
		count++
	}
	if object.valueType != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("value_type")
		stream.WriteString(*object.valueType)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalAddOnParameter reads a value of the 'add_on_parameter' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalAddOnParameter(source interface{}) (object *AddOnParameter, err error) {
	if source == http.NoBody {
		return
	}
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readAddOnParameter(iterator)
	err = iterator.Error
	return
}

// readAddOnParameter reads a value of the 'add_on_parameter' type from the given iterator.
func readAddOnParameter(iterator *jsoniter.Iterator) *AddOnParameter {
	object := &AddOnParameter{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == AddOnParameterLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "addon":
			value := readAddOn(iterator)
			object.addon = value
		case "default_value":
			value := iterator.ReadString()
			object.defaultValue = &value
		case "description":
			value := iterator.ReadString()
			object.description = &value
		case "editable":
			value := iterator.ReadBool()
			object.editable = &value
		case "enabled":
			value := iterator.ReadBool()
			object.enabled = &value
		case "name":
			value := iterator.ReadString()
			object.name = &value
		case "required":
			value := iterator.ReadBool()
			object.required = &value
		case "validation":
			value := iterator.ReadString()
			object.validation = &value
		case "value_type":
			value := iterator.ReadString()
			object.valueType = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
