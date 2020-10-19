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

// MachinePoolBuilder contains the data and logic needed to build 'machine_pool' objects.
//
// Representation of a machine pool in a cluster.
type MachinePoolBuilder struct {
	id                *string
	href              *string
	link              bool
	availabilityZones []string
	cluster           *ClusterBuilder
	instanceType      *string
	labels            map[string]string
	replicas          *int
}

// NewMachinePool creates a new builder of 'machine_pool' objects.
func NewMachinePool() *MachinePoolBuilder {
	return new(MachinePoolBuilder)
}

// ID sets the identifier of the object.
func (b *MachinePoolBuilder) ID(value string) *MachinePoolBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *MachinePoolBuilder) HREF(value string) *MachinePoolBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *MachinePoolBuilder) Link(value bool) *MachinePoolBuilder {
	b.link = value
	return b
}

// AvailabilityZones sets the value of the 'availability_zones' attribute to the given values.
//
//
func (b *MachinePoolBuilder) AvailabilityZones(values ...string) *MachinePoolBuilder {
	b.availabilityZones = make([]string, len(values))
	copy(b.availabilityZones, values)
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
func (b *MachinePoolBuilder) Cluster(value *ClusterBuilder) *MachinePoolBuilder {
	b.cluster = value
	return b
}

// InstanceType sets the value of the 'instance_type' attribute to the given value.
//
//
func (b *MachinePoolBuilder) InstanceType(value string) *MachinePoolBuilder {
	b.instanceType = &value
	return b
}

// Labels sets the value of the 'labels' attribute to the given value.
//
//
func (b *MachinePoolBuilder) Labels(value map[string]string) *MachinePoolBuilder {
	b.labels = value
	return b
}

// Replicas sets the value of the 'replicas' attribute to the given value.
//
//
func (b *MachinePoolBuilder) Replicas(value int) *MachinePoolBuilder {
	b.replicas = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *MachinePoolBuilder) Copy(object *MachinePool) *MachinePoolBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	if object.availabilityZones != nil {
		b.availabilityZones = make([]string, len(object.availabilityZones))
		copy(b.availabilityZones, object.availabilityZones)
	} else {
		b.availabilityZones = nil
	}
	if object.cluster != nil {
		b.cluster = NewCluster().Copy(object.cluster)
	} else {
		b.cluster = nil
	}
	b.instanceType = object.instanceType
	if len(object.labels) > 0 {
		b.labels = make(map[string]string)
		for k, v := range object.labels {
			b.labels[k] = v
		}
	} else {
		b.labels = nil
	}
	b.replicas = object.replicas
	return b
}

// Build creates a 'machine_pool' object using the configuration stored in the builder.
func (b *MachinePoolBuilder) Build() (object *MachinePool, err error) {
	object = new(MachinePool)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	if b.availabilityZones != nil {
		object.availabilityZones = make([]string, len(b.availabilityZones))
		copy(object.availabilityZones, b.availabilityZones)
	}
	if b.cluster != nil {
		object.cluster, err = b.cluster.Build()
		if err != nil {
			return
		}
	}
	object.instanceType = b.instanceType
	if b.labels != nil {
		object.labels = make(map[string]string)
		for k, v := range b.labels {
			object.labels[k] = v
		}
	}
	object.replicas = b.replicas
	return
}
