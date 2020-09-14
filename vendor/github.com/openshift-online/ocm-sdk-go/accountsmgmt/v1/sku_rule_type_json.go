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

package v1 // github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1

import (
	"io"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalSkuRule writes a value of the 'sku_rule' type to the given writer.
func MarshalSkuRule(object *SkuRule, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeSkuRule(object, stream)
	stream.Flush()
	return stream.Error
}

// writeSkuRule writes a value of the 'sku_rule' type to the given stream.
func writeSkuRule(object *SkuRule, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(SkuRuleLinkKind)
	} else {
		stream.WriteString(SkuRuleKind)
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
	if object.allowed != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("allowed")
		stream.WriteInt(*object.allowed)
		count++
	}
	if object.quotaId != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("quota_id")
		stream.WriteString(*object.quotaId)
		count++
	}
	if object.sku != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("sku")
		stream.WriteString(*object.sku)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalSkuRule reads a value of the 'sku_rule' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalSkuRule(source interface{}) (object *SkuRule, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readSkuRule(iterator)
	err = iterator.Error
	return
}

// readSkuRule reads a value of the 'sku_rule' type from the given iterator.
func readSkuRule(iterator *jsoniter.Iterator) *SkuRule {
	object := &SkuRule{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == SkuRuleLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "allowed":
			value := iterator.ReadInt()
			object.allowed = &value
		case "quota_id":
			value := iterator.ReadString()
			object.quotaId = &value
		case "sku":
			value := iterator.ReadString()
			object.sku = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
