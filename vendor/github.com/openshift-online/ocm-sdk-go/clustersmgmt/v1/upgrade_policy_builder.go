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
	time "time"
)

// UpgradePolicyBuilder contains the data and logic needed to build 'upgrade_policy' objects.
//
// Representation of an upgrade policy that can be set for a cluster.
type UpgradePolicyBuilder struct {
	id                   *string
	href                 *string
	link                 bool
	clusterID            *string
	nextRun              *time.Time
	nodeDrainGracePeriod *ValueBuilder
	schedule             *string
	scheduleType         *string
	upgradeType          *string
	version              *string
}

// NewUpgradePolicy creates a new builder of 'upgrade_policy' objects.
func NewUpgradePolicy() *UpgradePolicyBuilder {
	return new(UpgradePolicyBuilder)
}

// ID sets the identifier of the object.
func (b *UpgradePolicyBuilder) ID(value string) *UpgradePolicyBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *UpgradePolicyBuilder) HREF(value string) *UpgradePolicyBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *UpgradePolicyBuilder) Link(value bool) *UpgradePolicyBuilder {
	b.link = value
	return b
}

// ClusterID sets the value of the 'cluster_ID' attribute to the given value.
//
//
func (b *UpgradePolicyBuilder) ClusterID(value string) *UpgradePolicyBuilder {
	b.clusterID = &value
	return b
}

// NextRun sets the value of the 'next_run' attribute to the given value.
//
//
func (b *UpgradePolicyBuilder) NextRun(value time.Time) *UpgradePolicyBuilder {
	b.nextRun = &value
	return b
}

// NodeDrainGracePeriod sets the value of the 'node_drain_grace_period' attribute to the given value.
//
// Numeric value and the unit used to measure it.
//
// Units are not mandatory, and they're not specified for some resources. For
// resources that use bytes, the accepted units are:
//
// - 1 B = 1 byte
// - 1 KB = 10^3 bytes
// - 1 MB = 10^6 bytes
// - 1 GB = 10^9 bytes
// - 1 TB = 10^12 bytes
// - 1 PB = 10^15 bytes
//
// - 1 B = 1 byte
// - 1 KiB = 2^10 bytes
// - 1 MiB = 2^20 bytes
// - 1 GiB = 2^30 bytes
// - 1 TiB = 2^40 bytes
// - 1 PiB = 2^50 bytes
func (b *UpgradePolicyBuilder) NodeDrainGracePeriod(value *ValueBuilder) *UpgradePolicyBuilder {
	b.nodeDrainGracePeriod = value
	return b
}

// Schedule sets the value of the 'schedule' attribute to the given value.
//
//
func (b *UpgradePolicyBuilder) Schedule(value string) *UpgradePolicyBuilder {
	b.schedule = &value
	return b
}

// ScheduleType sets the value of the 'schedule_type' attribute to the given value.
//
//
func (b *UpgradePolicyBuilder) ScheduleType(value string) *UpgradePolicyBuilder {
	b.scheduleType = &value
	return b
}

// UpgradeType sets the value of the 'upgrade_type' attribute to the given value.
//
//
func (b *UpgradePolicyBuilder) UpgradeType(value string) *UpgradePolicyBuilder {
	b.upgradeType = &value
	return b
}

// Version sets the value of the 'version' attribute to the given value.
//
//
func (b *UpgradePolicyBuilder) Version(value string) *UpgradePolicyBuilder {
	b.version = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *UpgradePolicyBuilder) Copy(object *UpgradePolicy) *UpgradePolicyBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.clusterID = object.clusterID
	b.nextRun = object.nextRun
	if object.nodeDrainGracePeriod != nil {
		b.nodeDrainGracePeriod = NewValue().Copy(object.nodeDrainGracePeriod)
	} else {
		b.nodeDrainGracePeriod = nil
	}
	b.schedule = object.schedule
	b.scheduleType = object.scheduleType
	b.upgradeType = object.upgradeType
	b.version = object.version
	return b
}

// Build creates a 'upgrade_policy' object using the configuration stored in the builder.
func (b *UpgradePolicyBuilder) Build() (object *UpgradePolicy, err error) {
	object = new(UpgradePolicy)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	object.clusterID = b.clusterID
	object.nextRun = b.nextRun
	if b.nodeDrainGracePeriod != nil {
		object.nodeDrainGracePeriod, err = b.nodeDrainGracePeriod.Build()
		if err != nil {
			return
		}
	}
	object.schedule = b.schedule
	object.scheduleType = b.scheduleType
	object.upgradeType = b.upgradeType
	object.version = b.version
	return
}
