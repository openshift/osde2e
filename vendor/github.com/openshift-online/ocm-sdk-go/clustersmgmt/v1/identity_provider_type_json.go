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

// MarshalIdentityProvider writes a value of the 'identity_provider' type to the given writer.
func MarshalIdentityProvider(object *IdentityProvider, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeIdentityProvider(object, stream)
	stream.Flush()
	return stream.Error
}

// writeIdentityProvider writes a value of the 'identity_provider' type to the given stream.
func writeIdentityProvider(object *IdentityProvider, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(IdentityProviderLinkKind)
	} else {
		stream.WriteString(IdentityProviderKind)
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
	if object.ldap != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("ldap")
		writeLDAPIdentityProvider(object.ldap, stream)
		count++
	}
	if object.challenge != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("challenge")
		stream.WriteBool(*object.challenge)
		count++
	}
	if object.github != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("github")
		writeGithubIdentityProvider(object.github, stream)
		count++
	}
	if object.gitlab != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("gitlab")
		writeGitlabIdentityProvider(object.gitlab, stream)
		count++
	}
	if object.google != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("google")
		writeGoogleIdentityProvider(object.google, stream)
		count++
	}
	if object.login != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("login")
		stream.WriteBool(*object.login)
		count++
	}
	if object.mappingMethod != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("mapping_method")
		stream.WriteString(string(*object.mappingMethod))
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
	if object.openID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("open_id")
		writeOpenIDIdentityProvider(object.openID, stream)
		count++
	}
	if object.type_ != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("type")
		stream.WriteString(string(*object.type_))
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalIdentityProvider reads a value of the 'identity_provider' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalIdentityProvider(source interface{}) (object *IdentityProvider, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readIdentityProvider(iterator)
	err = iterator.Error
	return
}

// readIdentityProvider reads a value of the 'identity_provider' type from the given iterator.
func readIdentityProvider(iterator *jsoniter.Iterator) *IdentityProvider {
	object := &IdentityProvider{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == IdentityProviderLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "ldap":
			value := readLDAPIdentityProvider(iterator)
			object.ldap = value
		case "challenge":
			value := iterator.ReadBool()
			object.challenge = &value
		case "github":
			value := readGithubIdentityProvider(iterator)
			object.github = value
		case "gitlab":
			value := readGitlabIdentityProvider(iterator)
			object.gitlab = value
		case "google":
			value := readGoogleIdentityProvider(iterator)
			object.google = value
		case "login":
			value := iterator.ReadBool()
			object.login = &value
		case "mapping_method":
			text := iterator.ReadString()
			value := IdentityProviderMappingMethod(text)
			object.mappingMethod = &value
		case "name":
			value := iterator.ReadString()
			object.name = &value
		case "open_id":
			value := readOpenIDIdentityProvider(iterator)
			object.openID = value
		case "type":
			text := iterator.ReadString()
			value := IdentityProviderType(text)
			object.type_ = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
