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

// MarshalGCP writes a value of the 'GCP' type to the given writer.
func MarshalGCP(object *GCP, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeGCP(object, stream)
	stream.Flush()
	return stream.Error
}

// writeGCP writes a value of the 'GCP' type to the given stream.
func writeGCP(object *GCP, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if object.authURI != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("auth_uri")
		stream.WriteString(*object.authURI)
		count++
	}
	if object.authProviderX509CertURL != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("auth_provider_x509_cert_url")
		stream.WriteString(*object.authProviderX509CertURL)
		count++
	}
	if object.clientID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("client_id")
		stream.WriteString(*object.clientID)
		count++
	}
	if object.clientX509CertURL != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("client_x509_cert_url")
		stream.WriteString(*object.clientX509CertURL)
		count++
	}
	if object.clientEmail != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("client_email")
		stream.WriteString(*object.clientEmail)
		count++
	}
	if object.privateKey != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("private_key")
		stream.WriteString(*object.privateKey)
		count++
	}
	if object.privateKeyID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("private_key_id")
		stream.WriteString(*object.privateKeyID)
		count++
	}
	if object.projectID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("project_id")
		stream.WriteString(*object.projectID)
		count++
	}
	if object.tokenURI != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("token_uri")
		stream.WriteString(*object.tokenURI)
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

// UnmarshalGCP reads a value of the 'GCP' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalGCP(source interface{}) (object *GCP, err error) {
	if source == http.NoBody {
		return
	}
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readGCP(iterator)
	err = iterator.Error
	return
}

// readGCP reads a value of the 'GCP' type from the given iterator.
func readGCP(iterator *jsoniter.Iterator) *GCP {
	object := &GCP{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "auth_uri":
			value := iterator.ReadString()
			object.authURI = &value
		case "auth_provider_x509_cert_url":
			value := iterator.ReadString()
			object.authProviderX509CertURL = &value
		case "client_id":
			value := iterator.ReadString()
			object.clientID = &value
		case "client_x509_cert_url":
			value := iterator.ReadString()
			object.clientX509CertURL = &value
		case "client_email":
			value := iterator.ReadString()
			object.clientEmail = &value
		case "private_key":
			value := iterator.ReadString()
			object.privateKey = &value
		case "private_key_id":
			value := iterator.ReadString()
			object.privateKeyID = &value
		case "project_id":
			value := iterator.ReadString()
			object.projectID = &value
		case "token_uri":
			value := iterator.ReadString()
			object.tokenURI = &value
		case "type":
			value := iterator.ReadString()
			object.type_ = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
