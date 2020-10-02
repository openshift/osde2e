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

// MarshalProvisionShard writes a value of the 'provision_shard' type to the given writer.
func MarshalProvisionShard(object *ProvisionShard, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeProvisionShard(object, stream)
	stream.Flush()
	return stream.Error
}

// writeProvisionShard writes a value of the 'provision_shard' type to the given stream.
func writeProvisionShard(object *ProvisionShard, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(ProvisionShardLinkKind)
	} else {
		stream.WriteString(ProvisionShardKind)
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
	if object.awsAccountOperatorConfig != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("aws_account_operator_config")
		writeServerConfig(object.awsAccountOperatorConfig, stream)
		count++
	}
	if object.awsBaseDomain != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("aws_base_domain")
		stream.WriteString(*object.awsBaseDomain)
		count++
	}
	if object.gcpBaseDomain != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("gcp_base_domain")
		stream.WriteString(*object.gcpBaseDomain)
		count++
	}
	if object.gcpProjectOperator != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("gcp_project_operator")
		writeServerConfig(object.gcpProjectOperator, stream)
		count++
	}
	if object.hiveConfig != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("hive_config")
		writeServerConfig(object.hiveConfig, stream)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalProvisionShard reads a value of the 'provision_shard' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalProvisionShard(source interface{}) (object *ProvisionShard, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readProvisionShard(iterator)
	err = iterator.Error
	return
}

// readProvisionShard reads a value of the 'provision_shard' type from the given iterator.
func readProvisionShard(iterator *jsoniter.Iterator) *ProvisionShard {
	object := &ProvisionShard{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == ProvisionShardLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "aws_account_operator_config":
			value := readServerConfig(iterator)
			object.awsAccountOperatorConfig = value
		case "aws_base_domain":
			value := iterator.ReadString()
			object.awsBaseDomain = &value
		case "gcp_base_domain":
			value := iterator.ReadString()
			object.gcpBaseDomain = &value
		case "gcp_project_operator":
			value := readServerConfig(iterator)
			object.gcpProjectOperator = value
		case "hive_config":
			value := readServerConfig(iterator)
			object.hiveConfig = value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
