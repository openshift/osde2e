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

// MarshalSample writes a value of the 'sample' type to the given writer.
func MarshalSample(object *Sample, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeSample(object, stream)
	stream.Flush()
	return stream.Error
}

// writeSample writes a value of the 'sample' type to the given stream.
func writeSample(object *Sample, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if object.time != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("time")
		stream.WriteString((*object.time).Format(time.RFC3339))
		count++
	}
	if object.value != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("value")
		stream.WriteFloat64(*object.value)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalSample reads a value of the 'sample' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalSample(source interface{}) (object *Sample, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readSample(iterator)
	err = iterator.Error
	return
}

// readSample reads a value of the 'sample' type from the given iterator.
func readSample(iterator *jsoniter.Iterator) *Sample {
	object := &Sample{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "time":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.time = &value
		case "value":
			value := iterator.ReadFloat64()
			object.value = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
