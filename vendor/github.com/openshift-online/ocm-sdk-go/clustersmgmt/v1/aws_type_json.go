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

// MarshalAWS writes a value of the 'AWS' type to the given writer.
func MarshalAWS(object *AWS, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeAWS(object, stream)
	stream.Flush()
	return stream.Error
}

// writeAWS writes a value of the 'AWS' type to the given stream.
func writeAWS(object *AWS, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if object.accessKeyID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("access_key_id")
		stream.WriteString(*object.accessKeyID)
		count++
	}
	if object.accountID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("account_id")
		stream.WriteString(*object.accountID)
		count++
	}
	if object.secretAccessKey != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("secret_access_key")
		stream.WriteString(*object.secretAccessKey)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalAWS reads a value of the 'AWS' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalAWS(source interface{}) (object *AWS, err error) {
	if source == http.NoBody {
		return
	}
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readAWS(iterator)
	err = iterator.Error
	return
}

// readAWS reads a value of the 'AWS' type from the given iterator.
func readAWS(iterator *jsoniter.Iterator) *AWS {
	object := &AWS{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "access_key_id":
			value := iterator.ReadString()
			object.accessKeyID = &value
		case "account_id":
			value := iterator.ReadString()
			object.accountID = &value
		case "secret_access_key":
			value := iterator.ReadString()
			object.secretAccessKey = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
