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
	time "time"
)

// SubscriptionBuilder contains the data and logic needed to build 'subscription' objects.
//
//
type SubscriptionBuilder struct {
	id                *string
	href              *string
	link              bool
	clusterID         *string
	consumerUUID      *string
	cpuTotal          *int
	createdAt         *time.Time
	creator           *AccountBuilder
	displayName       *string
	externalClusterID *string
	labels            []*LabelBuilder
	lastReconcileDate *time.Time
	lastTelemetryDate *time.Time
	managed           *bool
	organizationID    *string
	plan              *PlanBuilder
	productBundle     *ProductBundleEnum
	serviceLevel      *ServiceLevelEnum
	socketTotal       *int
	status            *string
	supportLevel      *SupportLevelEnum
	systemUnits       *SystemUnitsEnum
	updatedAt         *time.Time
	usage             *UsageEnum
}

// NewSubscription creates a new builder of 'subscription' objects.
func NewSubscription() *SubscriptionBuilder {
	return new(SubscriptionBuilder)
}

// ID sets the identifier of the object.
func (b *SubscriptionBuilder) ID(value string) *SubscriptionBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *SubscriptionBuilder) HREF(value string) *SubscriptionBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *SubscriptionBuilder) Link(value bool) *SubscriptionBuilder {
	b.link = value
	return b
}

// ClusterID sets the value of the 'cluster_ID' attribute to the given value.
//
//
func (b *SubscriptionBuilder) ClusterID(value string) *SubscriptionBuilder {
	b.clusterID = &value
	return b
}

// ConsumerUUID sets the value of the 'consumer_UUID' attribute to the given value.
//
//
func (b *SubscriptionBuilder) ConsumerUUID(value string) *SubscriptionBuilder {
	b.consumerUUID = &value
	return b
}

// CpuTotal sets the value of the 'cpu_total' attribute to the given value.
//
//
func (b *SubscriptionBuilder) CpuTotal(value int) *SubscriptionBuilder {
	b.cpuTotal = &value
	return b
}

// CreatedAt sets the value of the 'created_at' attribute to the given value.
//
//
func (b *SubscriptionBuilder) CreatedAt(value time.Time) *SubscriptionBuilder {
	b.createdAt = &value
	return b
}

// Creator sets the value of the 'creator' attribute to the given value.
//
//
func (b *SubscriptionBuilder) Creator(value *AccountBuilder) *SubscriptionBuilder {
	b.creator = value
	return b
}

// DisplayName sets the value of the 'display_name' attribute to the given value.
//
//
func (b *SubscriptionBuilder) DisplayName(value string) *SubscriptionBuilder {
	b.displayName = &value
	return b
}

// ExternalClusterID sets the value of the 'external_cluster_ID' attribute to the given value.
//
//
func (b *SubscriptionBuilder) ExternalClusterID(value string) *SubscriptionBuilder {
	b.externalClusterID = &value
	return b
}

// Labels sets the value of the 'labels' attribute to the given values.
//
//
func (b *SubscriptionBuilder) Labels(values ...*LabelBuilder) *SubscriptionBuilder {
	b.labels = make([]*LabelBuilder, len(values))
	copy(b.labels, values)
	return b
}

// LastReconcileDate sets the value of the 'last_reconcile_date' attribute to the given value.
//
//
func (b *SubscriptionBuilder) LastReconcileDate(value time.Time) *SubscriptionBuilder {
	b.lastReconcileDate = &value
	return b
}

// LastTelemetryDate sets the value of the 'last_telemetry_date' attribute to the given value.
//
//
func (b *SubscriptionBuilder) LastTelemetryDate(value time.Time) *SubscriptionBuilder {
	b.lastTelemetryDate = &value
	return b
}

// Managed sets the value of the 'managed' attribute to the given value.
//
//
func (b *SubscriptionBuilder) Managed(value bool) *SubscriptionBuilder {
	b.managed = &value
	return b
}

// OrganizationID sets the value of the 'organization_ID' attribute to the given value.
//
//
func (b *SubscriptionBuilder) OrganizationID(value string) *SubscriptionBuilder {
	b.organizationID = &value
	return b
}

// Plan sets the value of the 'plan' attribute to the given value.
//
//
func (b *SubscriptionBuilder) Plan(value *PlanBuilder) *SubscriptionBuilder {
	b.plan = value
	return b
}

// ProductBundle sets the value of the 'product_bundle' attribute to the given value.
//
// Usage of Subscription.
func (b *SubscriptionBuilder) ProductBundle(value ProductBundleEnum) *SubscriptionBuilder {
	b.productBundle = &value
	return b
}

// ServiceLevel sets the value of the 'service_level' attribute to the given value.
//
// Service Level of Subscription.
func (b *SubscriptionBuilder) ServiceLevel(value ServiceLevelEnum) *SubscriptionBuilder {
	b.serviceLevel = &value
	return b
}

// SocketTotal sets the value of the 'socket_total' attribute to the given value.
//
//
func (b *SubscriptionBuilder) SocketTotal(value int) *SubscriptionBuilder {
	b.socketTotal = &value
	return b
}

// Status sets the value of the 'status' attribute to the given value.
//
//
func (b *SubscriptionBuilder) Status(value string) *SubscriptionBuilder {
	b.status = &value
	return b
}

// SupportLevel sets the value of the 'support_level' attribute to the given value.
//
// Support Level of Subscription.
func (b *SubscriptionBuilder) SupportLevel(value SupportLevelEnum) *SubscriptionBuilder {
	b.supportLevel = &value
	return b
}

// SystemUnits sets the value of the 'system_units' attribute to the given value.
//
// Usage of Subscription.
func (b *SubscriptionBuilder) SystemUnits(value SystemUnitsEnum) *SubscriptionBuilder {
	b.systemUnits = &value
	return b
}

// UpdatedAt sets the value of the 'updated_at' attribute to the given value.
//
//
func (b *SubscriptionBuilder) UpdatedAt(value time.Time) *SubscriptionBuilder {
	b.updatedAt = &value
	return b
}

// Usage sets the value of the 'usage' attribute to the given value.
//
// Usage of Subscription.
func (b *SubscriptionBuilder) Usage(value UsageEnum) *SubscriptionBuilder {
	b.usage = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *SubscriptionBuilder) Copy(object *Subscription) *SubscriptionBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.clusterID = object.clusterID
	b.consumerUUID = object.consumerUUID
	b.cpuTotal = object.cpuTotal
	b.createdAt = object.createdAt
	if object.creator != nil {
		b.creator = NewAccount().Copy(object.creator)
	} else {
		b.creator = nil
	}
	b.displayName = object.displayName
	b.externalClusterID = object.externalClusterID
	if object.labels != nil {
		b.labels = make([]*LabelBuilder, len(object.labels))
		for i, v := range object.labels {
			b.labels[i] = NewLabel().Copy(v)
		}
	} else {
		b.labels = nil
	}
	b.lastReconcileDate = object.lastReconcileDate
	b.lastTelemetryDate = object.lastTelemetryDate
	b.managed = object.managed
	b.organizationID = object.organizationID
	if object.plan != nil {
		b.plan = NewPlan().Copy(object.plan)
	} else {
		b.plan = nil
	}
	b.productBundle = object.productBundle
	b.serviceLevel = object.serviceLevel
	b.socketTotal = object.socketTotal
	b.status = object.status
	b.supportLevel = object.supportLevel
	b.systemUnits = object.systemUnits
	b.updatedAt = object.updatedAt
	b.usage = object.usage
	return b
}

// Build creates a 'subscription' object using the configuration stored in the builder.
func (b *SubscriptionBuilder) Build() (object *Subscription, err error) {
	object = new(Subscription)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	object.clusterID = b.clusterID
	object.consumerUUID = b.consumerUUID
	object.cpuTotal = b.cpuTotal
	object.createdAt = b.createdAt
	if b.creator != nil {
		object.creator, err = b.creator.Build()
		if err != nil {
			return
		}
	}
	object.displayName = b.displayName
	object.externalClusterID = b.externalClusterID
	if b.labels != nil {
		object.labels = make([]*Label, len(b.labels))
		for i, v := range b.labels {
			object.labels[i], err = v.Build()
			if err != nil {
				return
			}
		}
	}
	object.lastReconcileDate = b.lastReconcileDate
	object.lastTelemetryDate = b.lastTelemetryDate
	object.managed = b.managed
	object.organizationID = b.organizationID
	if b.plan != nil {
		object.plan, err = b.plan.Build()
		if err != nil {
			return
		}
	}
	object.productBundle = b.productBundle
	object.serviceLevel = b.serviceLevel
	object.socketTotal = b.socketTotal
	object.status = b.status
	object.supportLevel = b.supportLevel
	object.systemUnits = b.systemUnits
	object.updatedAt = b.updatedAt
	object.usage = b.usage
	return
}
