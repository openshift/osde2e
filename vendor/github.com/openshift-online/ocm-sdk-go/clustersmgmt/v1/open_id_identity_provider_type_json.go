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
	"sort"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalOpenIDIdentityProvider writes a value of the 'open_ID_identity_provider' type to the given writer.
func MarshalOpenIDIdentityProvider(object *OpenIDIdentityProvider, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeOpenIDIdentityProvider(object, stream)
	stream.Flush()
	return stream.Error
}

// writeOpenIDIdentityProvider writes a value of the 'open_ID_identity_provider' type to the given stream.
func writeOpenIDIdentityProvider(object *OpenIDIdentityProvider, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if object.ca != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("ca")
		stream.WriteString(*object.ca)
		count++
	}
	if object.claims != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("claims")
		writeOpenIDClaims(object.claims, stream)
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
	if object.clientSecret != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("client_secret")
		stream.WriteString(*object.clientSecret)
		count++
	}
	if object.extraAuthorizeParameters != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("extra_authorize_parameters")
		stream.WriteObjectStart()
		keys := make([]string, len(object.extraAuthorizeParameters))
		i := 0
		for key := range object.extraAuthorizeParameters {
			keys[i] = key
			i++
		}
		sort.Strings(keys)
		for i, key := range keys {
			if i > 0 {
				stream.WriteMore()
			}
			item := object.extraAuthorizeParameters[key]
			stream.WriteObjectField(key)
			stream.WriteString(item)
		}
		stream.WriteObjectEnd()
		count++
	}
	if object.extraScopes != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("extra_scopes")
		writeStringList(object.extraScopes, stream)
		count++
	}
	if object.issuer != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("issuer")
		stream.WriteString(*object.issuer)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalOpenIDIdentityProvider reads a value of the 'open_ID_identity_provider' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalOpenIDIdentityProvider(source interface{}) (object *OpenIDIdentityProvider, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readOpenIDIdentityProvider(iterator)
	err = iterator.Error
	return
}

// readOpenIDIdentityProvider reads a value of the 'open_ID_identity_provider' type from the given iterator.
func readOpenIDIdentityProvider(iterator *jsoniter.Iterator) *OpenIDIdentityProvider {
	object := &OpenIDIdentityProvider{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "ca":
			value := iterator.ReadString()
			object.ca = &value
		case "claims":
			value := readOpenIDClaims(iterator)
			object.claims = value
		case "client_id":
			value := iterator.ReadString()
			object.clientID = &value
		case "client_secret":
			value := iterator.ReadString()
			object.clientSecret = &value
		case "extra_authorize_parameters":
			value := map[string]string{}
			for {
				key := iterator.ReadObject()
				if key == "" {
					break
				}
				item := iterator.ReadString()
				value[key] = item
			}
			object.extraAuthorizeParameters = value
		case "extra_scopes":
			value := readStringList(iterator)
			object.extraScopes = value
		case "issuer":
			value := iterator.ReadString()
			object.issuer = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
