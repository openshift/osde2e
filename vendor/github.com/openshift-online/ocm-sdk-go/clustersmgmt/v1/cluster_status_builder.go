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

// ClusterStatusBuilder contains the data and logic needed to build 'cluster_status' objects.
//
// Detailed status of a cluster.
type ClusterStatusBuilder struct {
	id          *string
	href        *string
	link        bool
	description *string
	state       *ClusterState
}

// NewClusterStatus creates a new builder of 'cluster_status' objects.
func NewClusterStatus() *ClusterStatusBuilder {
	return new(ClusterStatusBuilder)
}

// ID sets the identifier of the object.
func (b *ClusterStatusBuilder) ID(value string) *ClusterStatusBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *ClusterStatusBuilder) HREF(value string) *ClusterStatusBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *ClusterStatusBuilder) Link(value bool) *ClusterStatusBuilder {
	b.link = value
	return b
}

// Description sets the value of the 'description' attribute to the given value.
//
//
func (b *ClusterStatusBuilder) Description(value string) *ClusterStatusBuilder {
	b.description = &value
	return b
}

// State sets the value of the 'state' attribute to the given value.
//
// Overall state of a cluster.
func (b *ClusterStatusBuilder) State(value ClusterState) *ClusterStatusBuilder {
	b.state = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *ClusterStatusBuilder) Copy(object *ClusterStatus) *ClusterStatusBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.description = object.description
	b.state = object.state
	return b
}

// Build creates a 'cluster_status' object using the configuration stored in the builder.
func (b *ClusterStatusBuilder) Build() (object *ClusterStatus, err error) {
	object = new(ClusterStatus)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	object.description = b.description
	object.state = b.state
	return
}
