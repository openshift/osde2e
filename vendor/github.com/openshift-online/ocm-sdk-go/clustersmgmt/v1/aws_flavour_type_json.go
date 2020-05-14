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

// MarshalAWSFlavour writes a value of the 'AWS_flavour' type to the given writer.
func MarshalAWSFlavour(object *AWSFlavour, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeAWSFlavour(object, stream)
	stream.Flush()
	return stream.Error
}

// writeAWSFlavour writes a value of the 'AWS_flavour' type to the given stream.
func writeAWSFlavour(object *AWSFlavour, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if object.computeInstanceType != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("compute_instance_type")
		stream.WriteString(*object.computeInstanceType)
		count++
	}
	if object.infraInstanceType != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("infra_instance_type")
		stream.WriteString(*object.infraInstanceType)
		count++
	}
	if object.infraVolume != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("infra_volume")
		writeAWSVolume(object.infraVolume, stream)
		count++
	}
	if object.masterInstanceType != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("master_instance_type")
		stream.WriteString(*object.masterInstanceType)
		count++
	}
	if object.masterVolume != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("master_volume")
		writeAWSVolume(object.masterVolume, stream)
		count++
	}
	if object.workerVolume != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("worker_volume")
		writeAWSVolume(object.workerVolume, stream)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalAWSFlavour reads a value of the 'AWS_flavour' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalAWSFlavour(source interface{}) (object *AWSFlavour, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readAWSFlavour(iterator)
	err = iterator.Error
	return
}

// readAWSFlavour reads a value of the 'AWS_flavour' type from the given iterator.
func readAWSFlavour(iterator *jsoniter.Iterator) *AWSFlavour {
	object := &AWSFlavour{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "compute_instance_type":
			value := iterator.ReadString()
			object.computeInstanceType = &value
		case "infra_instance_type":
			value := iterator.ReadString()
			object.infraInstanceType = &value
		case "infra_volume":
			value := readAWSVolume(iterator)
			object.infraVolume = value
		case "master_instance_type":
			value := iterator.ReadString()
			object.masterInstanceType = &value
		case "master_volume":
			value := readAWSVolume(iterator)
			object.masterVolume = value
		case "worker_volume":
			value := readAWSVolume(iterator)
			object.workerVolume = value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
