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

// MarshalLDAPIdentityProvider writes a value of the 'LDAP_identity_provider' type to the given writer.
func MarshalLDAPIdentityProvider(object *LDAPIdentityProvider, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeLDAPIdentityProvider(object, stream)
	stream.Flush()
	return stream.Error
}

// writeLDAPIdentityProvider writes a value of the 'LDAP_identity_provider' type to the given stream.
func writeLDAPIdentityProvider(object *LDAPIdentityProvider, stream *jsoniter.Stream) {
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
	if object.url != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("url")
		stream.WriteString(*object.url)
		count++
	}
	if object.attributes != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("attributes")
		writeLDAPAttributes(object.attributes, stream)
		count++
	}
	if object.bindDN != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("bind_dn")
		stream.WriteString(*object.bindDN)
		count++
	}
	if object.bindPassword != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("bind_password")
		stream.WriteString(*object.bindPassword)
		count++
	}
	if object.insecure != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("insecure")
		stream.WriteBool(*object.insecure)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalLDAPIdentityProvider reads a value of the 'LDAP_identity_provider' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalLDAPIdentityProvider(source interface{}) (object *LDAPIdentityProvider, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readLDAPIdentityProvider(iterator)
	err = iterator.Error
	return
}

// readLDAPIdentityProvider reads a value of the 'LDAP_identity_provider' type from the given iterator.
func readLDAPIdentityProvider(iterator *jsoniter.Iterator) *LDAPIdentityProvider {
	object := &LDAPIdentityProvider{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "ca":
			value := iterator.ReadString()
			object.ca = &value
		case "url":
			value := iterator.ReadString()
			object.url = &value
		case "attributes":
			value := readLDAPAttributes(iterator)
			object.attributes = value
		case "bind_dn":
			value := iterator.ReadString()
			object.bindDN = &value
		case "bind_password":
			value := iterator.ReadString()
			object.bindPassword = &value
		case "insecure":
			value := iterator.ReadBool()
			object.insecure = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
