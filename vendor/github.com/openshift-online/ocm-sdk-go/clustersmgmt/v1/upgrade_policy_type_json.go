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
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalUpgradePolicy writes a value of the 'upgrade_policy' type to the given writer.
func MarshalUpgradePolicy(object *UpgradePolicy, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeUpgradePolicy(object, stream)
	stream.Flush()
	return stream.Error
}

// writeUpgradePolicy writes a value of the 'upgrade_policy' type to the given stream.
func writeUpgradePolicy(object *UpgradePolicy, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(UpgradePolicyLinkKind)
	} else {
		stream.WriteString(UpgradePolicyKind)
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
	if object.clusterID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cluster_id")
		stream.WriteString(*object.clusterID)
		count++
	}
	if object.nextRun != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("next_run")
		stream.WriteString((*object.nextRun).Format(time.RFC3339))
		count++
	}
	if object.nodeDrainGracePeriod != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("node_drain_grace_period")
		writeValue(object.nodeDrainGracePeriod, stream)
		count++
	}
	if object.schedule != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("schedule")
		stream.WriteString(*object.schedule)
		count++
	}
	if object.scheduleType != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("schedule_type")
		stream.WriteString(*object.scheduleType)
		count++
	}
	if object.upgradeType != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("upgrade_type")
		stream.WriteString(*object.upgradeType)
		count++
	}
	if object.version != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("version")
		stream.WriteString(*object.version)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalUpgradePolicy reads a value of the 'upgrade_policy' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalUpgradePolicy(source interface{}) (object *UpgradePolicy, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readUpgradePolicy(iterator)
	err = iterator.Error
	return
}

// readUpgradePolicy reads a value of the 'upgrade_policy' type from the given iterator.
func readUpgradePolicy(iterator *jsoniter.Iterator) *UpgradePolicy {
	object := &UpgradePolicy{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == UpgradePolicyLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "cluster_id":
			value := iterator.ReadString()
			object.clusterID = &value
		case "next_run":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.nextRun = &value
		case "node_drain_grace_period":
			value := readValue(iterator)
			object.nodeDrainGracePeriod = value
		case "schedule":
			value := iterator.ReadString()
			object.schedule = &value
		case "schedule_type":
			value := iterator.ReadString()
			object.scheduleType = &value
		case "upgrade_type":
			value := iterator.ReadString()
			object.upgradeType = &value
		case "version":
			value := iterator.ReadString()
			object.version = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
