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

// MarshalFlavour writes a value of the 'flavour' type to the given writer.
func MarshalFlavour(object *Flavour, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeFlavour(object, stream)
	stream.Flush()
	return stream.Error
}

// writeFlavour writes a value of the 'flavour' type to the given stream.
func writeFlavour(object *Flavour, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(FlavourLinkKind)
	} else {
		stream.WriteString(FlavourKind)
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
	if object.aws != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("aws")
		writeAWSFlavour(object.aws, stream)
		count++
	}
	if object.gcp != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("gcp")
		writeGCPFlavour(object.gcp, stream)
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
	if object.network != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("network")
		writeNetwork(object.network, stream)
		count++
	}
	if object.nodes != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("nodes")
		writeFlavourNodes(object.nodes, stream)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalFlavour reads a value of the 'flavour' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalFlavour(source interface{}) (object *Flavour, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readFlavour(iterator)
	err = iterator.Error
	return
}

// readFlavour reads a value of the 'flavour' type from the given iterator.
func readFlavour(iterator *jsoniter.Iterator) *Flavour {
	object := &Flavour{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == FlavourLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "aws":
			value := readAWSFlavour(iterator)
			object.aws = value
		case "gcp":
			value := readGCPFlavour(iterator)
			object.gcp = value
		case "name":
			value := iterator.ReadString()
			object.name = &value
		case "network":
			value := readNetwork(iterator)
			object.network = value
		case "nodes":
			value := readFlavourNodes(iterator)
			object.nodes = value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
