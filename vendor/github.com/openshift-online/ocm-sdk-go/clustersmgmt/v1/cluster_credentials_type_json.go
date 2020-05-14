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

// MarshalClusterCredentials writes a value of the 'cluster_credentials' type to the given writer.
func MarshalClusterCredentials(object *ClusterCredentials, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeClusterCredentials(object, stream)
	stream.Flush()
	return stream.Error
}

// writeClusterCredentials writes a value of the 'cluster_credentials' type to the given stream.
func writeClusterCredentials(object *ClusterCredentials, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(ClusterCredentialsLinkKind)
	} else {
		stream.WriteString(ClusterCredentialsKind)
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
	if object.ssh != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("ssh")
		writeSSHCredentials(object.ssh, stream)
		count++
	}
	if object.admin != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("admin")
		writeAdminCredentials(object.admin, stream)
		count++
	}
	if object.kubeconfig != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("kubeconfig")
		stream.WriteString(*object.kubeconfig)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalClusterCredentials reads a value of the 'cluster_credentials' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalClusterCredentials(source interface{}) (object *ClusterCredentials, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readClusterCredentials(iterator)
	err = iterator.Error
	return
}

// readClusterCredentials reads a value of the 'cluster_credentials' type from the given iterator.
func readClusterCredentials(iterator *jsoniter.Iterator) *ClusterCredentials {
	object := &ClusterCredentials{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == ClusterCredentialsLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "ssh":
			value := readSSHCredentials(iterator)
			object.ssh = value
		case "admin":
			value := readAdminCredentials(iterator)
			object.admin = value
		case "kubeconfig":
			value := iterator.ReadString()
			object.kubeconfig = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
