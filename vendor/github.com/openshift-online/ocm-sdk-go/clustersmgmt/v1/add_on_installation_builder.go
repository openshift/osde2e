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

// AddOnInstallationBuilder contains the data and logic needed to build 'add_on_installation' objects.
//
// Representation of an add-on installation in a cluster.
type AddOnInstallationBuilder struct {
	id                *string
	href              *string
	link              bool
	addon             *AddOnBuilder
	cluster           *ClusterBuilder
	creationTimestamp *time.Time
	operatorVersion   *string
	state             *AddOnInstallationState
	stateDescription  *string
	updatedTimestamp  *time.Time
}

// NewAddOnInstallation creates a new builder of 'add_on_installation' objects.
func NewAddOnInstallation() *AddOnInstallationBuilder {
	return new(AddOnInstallationBuilder)
}

// ID sets the identifier of the object.
func (b *AddOnInstallationBuilder) ID(value string) *AddOnInstallationBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *AddOnInstallationBuilder) HREF(value string) *AddOnInstallationBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *AddOnInstallationBuilder) Link(value bool) *AddOnInstallationBuilder {
	b.link = value
	return b
}

// Addon sets the value of the 'addon' attribute to the given value.
//
// Representation of an add-on that can be installed in a cluster.
func (b *AddOnInstallationBuilder) Addon(value *AddOnBuilder) *AddOnInstallationBuilder {
	b.addon = value
	return b
}

// Cluster sets the value of the 'cluster' attribute to the given value.
//
// Definition of an _OpenShift_ cluster.
//
// The `cloud_provider` attribute is a reference to the cloud provider. When a
// cluster is retrieved it will be a link to the cloud provider, containing only
// the kind, id and href attributes:
//
// [source,json]
// ----
// {
//   "cloud_provider": {
//     "kind": "CloudProviderLink",
//     "id": "123",
//     "href": "/api/clusters_mgmt/v1/cloud_providers/123"
//   }
// }
// ----
//
// When a cluster is created this is optional, and if used it should contain the
// identifier of the cloud provider to use:
//
// [source,json]
// ----
// {
//   "cloud_provider": {
//     "id": "123",
//   }
// }
// ----
//
// If not included, then the cluster will be created using the default cloud
// provider, which is currently Amazon Web Services.
//
// The region attribute is mandatory when a cluster is created.
//
// The `aws.access_key_id`, `aws.secret_access_key` and `dns.base_domain`
// attributes are mandatory when creation a cluster with your own Amazon Web
// Services account.
func (b *AddOnInstallationBuilder) Cluster(value *ClusterBuilder) *AddOnInstallationBuilder {
	b.cluster = value
	return b
}

// CreationTimestamp sets the value of the 'creation_timestamp' attribute to the given value.
//
//
func (b *AddOnInstallationBuilder) CreationTimestamp(value time.Time) *AddOnInstallationBuilder {
	b.creationTimestamp = &value
	return b
}

// OperatorVersion sets the value of the 'operator_version' attribute to the given value.
//
//
func (b *AddOnInstallationBuilder) OperatorVersion(value string) *AddOnInstallationBuilder {
	b.operatorVersion = &value
	return b
}

// State sets the value of the 'state' attribute to the given value.
//
// Representation of an add-on installation State field.
func (b *AddOnInstallationBuilder) State(value AddOnInstallationState) *AddOnInstallationBuilder {
	b.state = &value
	return b
}

// StateDescription sets the value of the 'state_description' attribute to the given value.
//
//
func (b *AddOnInstallationBuilder) StateDescription(value string) *AddOnInstallationBuilder {
	b.stateDescription = &value
	return b
}

// UpdatedTimestamp sets the value of the 'updated_timestamp' attribute to the given value.
//
//
func (b *AddOnInstallationBuilder) UpdatedTimestamp(value time.Time) *AddOnInstallationBuilder {
	b.updatedTimestamp = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *AddOnInstallationBuilder) Copy(object *AddOnInstallation) *AddOnInstallationBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	if object.addon != nil {
		b.addon = NewAddOn().Copy(object.addon)
	} else {
		b.addon = nil
	}
	if object.cluster != nil {
		b.cluster = NewCluster().Copy(object.cluster)
	} else {
		b.cluster = nil
	}
	b.creationTimestamp = object.creationTimestamp
	b.operatorVersion = object.operatorVersion
	b.state = object.state
	b.stateDescription = object.stateDescription
	b.updatedTimestamp = object.updatedTimestamp
	return b
}

// Build creates a 'add_on_installation' object using the configuration stored in the builder.
func (b *AddOnInstallationBuilder) Build() (object *AddOnInstallation, err error) {
	object = new(AddOnInstallation)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	if b.addon != nil {
		object.addon, err = b.addon.Build()
		if err != nil {
			return
		}
	}
	if b.cluster != nil {
		object.cluster, err = b.cluster.Build()
		if err != nil {
			return
		}
	}
	object.creationTimestamp = b.creationTimestamp
	object.operatorVersion = b.operatorVersion
	object.state = b.state
	object.stateDescription = b.stateDescription
	object.updatedTimestamp = b.updatedTimestamp
	return
}
