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

// MachinePoolAutoscalingBuilder contains the data and logic needed to build 'machine_pool_autoscaling' objects.
//
// Representation of a autoscaling in a machine pool.
type MachinePoolAutoscalingBuilder struct {
	id          *string
	href        *string
	link        bool
	maxReplicas *int
	minReplicas *int
}

// NewMachinePoolAutoscaling creates a new builder of 'machine_pool_autoscaling' objects.
func NewMachinePoolAutoscaling() *MachinePoolAutoscalingBuilder {
	return new(MachinePoolAutoscalingBuilder)
}

// ID sets the identifier of the object.
func (b *MachinePoolAutoscalingBuilder) ID(value string) *MachinePoolAutoscalingBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *MachinePoolAutoscalingBuilder) HREF(value string) *MachinePoolAutoscalingBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *MachinePoolAutoscalingBuilder) Link(value bool) *MachinePoolAutoscalingBuilder {
	b.link = value
	return b
}

// MaxReplicas sets the value of the 'max_replicas' attribute to the given value.
//
//
func (b *MachinePoolAutoscalingBuilder) MaxReplicas(value int) *MachinePoolAutoscalingBuilder {
	b.maxReplicas = &value
	return b
}

// MinReplicas sets the value of the 'min_replicas' attribute to the given value.
//
//
func (b *MachinePoolAutoscalingBuilder) MinReplicas(value int) *MachinePoolAutoscalingBuilder {
	b.minReplicas = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *MachinePoolAutoscalingBuilder) Copy(object *MachinePoolAutoscaling) *MachinePoolAutoscalingBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.maxReplicas = object.maxReplicas
	b.minReplicas = object.minReplicas
	return b
}

// Build creates a 'machine_pool_autoscaling' object using the configuration stored in the builder.
func (b *MachinePoolAutoscalingBuilder) Build() (object *MachinePoolAutoscaling, err error) {
	object = new(MachinePoolAutoscaling)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	object.maxReplicas = b.maxReplicas
	object.minReplicas = b.minReplicas
	return
}
