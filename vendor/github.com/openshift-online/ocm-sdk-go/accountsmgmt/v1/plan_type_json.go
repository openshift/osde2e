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
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalPlan writes a value of the 'plan' type to the given writer.
func MarshalPlan(object *Plan, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writePlan(object, stream)
	stream.Flush()
	return stream.Error
}

// writePlan writes a value of the 'plan' type to the given stream.
func writePlan(object *Plan, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(PlanLinkKind)
	} else {
		stream.WriteString(PlanKind)
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
	if object.name != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("name")
		stream.WriteString(*object.name)
		count++
	}
	if object.type_ != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("type")
		stream.WriteString(*object.type_)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalPlan reads a value of the 'plan' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalPlan(source interface{}) (object *Plan, err error) {
	if source == http.NoBody {
		return
	}
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readPlan(iterator)
	err = iterator.Error
	return
}

// readPlan reads a value of the 'plan' type from the given iterator.
func readPlan(iterator *jsoniter.Iterator) *Plan {
	object := &Plan{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == PlanLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "name":
			value := iterator.ReadString()
			object.name = &value
		case "type":
			value := iterator.ReadString()
			object.type_ = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
