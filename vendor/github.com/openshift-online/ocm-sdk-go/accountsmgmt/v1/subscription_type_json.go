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
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalSubscription writes a value of the 'subscription' type to the given writer.
func MarshalSubscription(object *Subscription, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeSubscription(object, stream)
	stream.Flush()
	return stream.Error
}

// writeSubscription writes a value of the 'subscription' type to the given stream.
func writeSubscription(object *Subscription, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(SubscriptionLinkKind)
	} else {
		stream.WriteString(SubscriptionKind)
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
	if object.consumerUUID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("consumer_uuid")
		stream.WriteString(*object.consumerUUID)
		count++
	}
	if object.cpuTotal != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cpu_total")
		stream.WriteInt(*object.cpuTotal)
		count++
	}
	if object.createdAt != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("created_at")
		stream.WriteString((*object.createdAt).Format(time.RFC3339))
		count++
	}
	if object.creator != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("creator")
		writeAccount(object.creator, stream)
		count++
	}
	if object.displayName != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("display_name")
		stream.WriteString(*object.displayName)
		count++
	}
	if object.externalClusterID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("external_cluster_id")
		stream.WriteString(*object.externalClusterID)
		count++
	}
	if object.labels != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("labels")
		writeLabelList(object.labels, stream)
		count++
	}
	if object.lastReconcileDate != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("last_reconcile_date")
		stream.WriteString((*object.lastReconcileDate).Format(time.RFC3339))
		count++
	}
	if object.lastTelemetryDate != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("last_telemetry_date")
		stream.WriteString((*object.lastTelemetryDate).Format(time.RFC3339))
		count++
	}
	if object.managed != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("managed")
		stream.WriteBool(*object.managed)
		count++
	}
	if object.organizationID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("organization_id")
		stream.WriteString(*object.organizationID)
		count++
	}
	if object.plan != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("plan")
		writePlan(object.plan, stream)
		count++
	}
	if object.productBundle != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("product_bundle")
		stream.WriteString(string(*object.productBundle))
		count++
	}
	if object.serviceLevel != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("service_level")
		stream.WriteString(string(*object.serviceLevel))
		count++
	}
	if object.socketTotal != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("socket_total")
		stream.WriteInt(*object.socketTotal)
		count++
	}
	if object.status != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("status")
		stream.WriteString(*object.status)
		count++
	}
	if object.supportLevel != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("support_level")
		stream.WriteString(string(*object.supportLevel))
		count++
	}
	if object.systemUnits != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("system_units")
		stream.WriteString(string(*object.systemUnits))
		count++
	}
	if object.updatedAt != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("updated_at")
		stream.WriteString((*object.updatedAt).Format(time.RFC3339))
		count++
	}
	if object.usage != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("usage")
		stream.WriteString(string(*object.usage))
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalSubscription reads a value of the 'subscription' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalSubscription(source interface{}) (object *Subscription, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readSubscription(iterator)
	err = iterator.Error
	return
}

// readSubscription reads a value of the 'subscription' type from the given iterator.
func readSubscription(iterator *jsoniter.Iterator) *Subscription {
	object := &Subscription{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == SubscriptionLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "cluster_id":
			value := iterator.ReadString()
			object.clusterID = &value
		case "consumer_uuid":
			value := iterator.ReadString()
			object.consumerUUID = &value
		case "cpu_total":
			value := iterator.ReadInt()
			object.cpuTotal = &value
		case "created_at":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.createdAt = &value
		case "creator":
			value := readAccount(iterator)
			object.creator = value
		case "display_name":
			value := iterator.ReadString()
			object.displayName = &value
		case "external_cluster_id":
			value := iterator.ReadString()
			object.externalClusterID = &value
		case "labels":
			value := readLabelList(iterator)
			object.labels = value
		case "last_reconcile_date":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.lastReconcileDate = &value
		case "last_telemetry_date":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.lastTelemetryDate = &value
		case "managed":
			value := iterator.ReadBool()
			object.managed = &value
		case "organization_id":
			value := iterator.ReadString()
			object.organizationID = &value
		case "plan":
			value := readPlan(iterator)
			object.plan = value
		case "product_bundle":
			text := iterator.ReadString()
			value := ProductBundleEnum(text)
			object.productBundle = &value
		case "service_level":
			text := iterator.ReadString()
			value := ServiceLevelEnum(text)
			object.serviceLevel = &value
		case "socket_total":
			value := iterator.ReadInt()
			object.socketTotal = &value
		case "status":
			value := iterator.ReadString()
			object.status = &value
		case "support_level":
			text := iterator.ReadString()
			value := SupportLevelEnum(text)
			object.supportLevel = &value
		case "system_units":
			text := iterator.ReadString()
			value := SystemUnitsEnum(text)
			object.systemUnits = &value
		case "updated_at":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.updatedAt = &value
		case "usage":
			text := iterator.ReadString()
			value := UsageEnum(text)
			object.usage = &value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
