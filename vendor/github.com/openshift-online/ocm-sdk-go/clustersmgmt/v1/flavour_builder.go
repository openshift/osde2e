/*
Copyright (c) 2019 Red Hat, Inc.

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

// FlavourBuilder contains the data and logic needed to build 'flavour' objects.
//
// Set of predefined properties of a cluster. For example, a _huge_ flavour can be a cluster
// with 10 infra nodes and 1000 compute nodes.
type FlavourBuilder struct {
	id                  *string
	href                *string
	link                bool
	aws                 *AWSFlavourBuilder
	computeInstanceType *string
	infraInstanceType   *string
	masterInstanceType  *string
	name                *string
	network             *NetworkBuilder
	nodes               *FlavourNodesBuilder
}

// NewFlavour creates a new builder of 'flavour' objects.
func NewFlavour() *FlavourBuilder {
	return new(FlavourBuilder)
}

// ID sets the identifier of the object.
func (b *FlavourBuilder) ID(value string) *FlavourBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *FlavourBuilder) HREF(value string) *FlavourBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *FlavourBuilder) Link(value bool) *FlavourBuilder {
	b.link = value
	return b
}

// AWS sets the value of the 'AWS' attribute
// to the given value.
//
// Volume specification for different classes of nodes inside a flavour.
func (b *FlavourBuilder) AWS(value *AWSFlavourBuilder) *FlavourBuilder {
	b.aws = value
	return b
}

// ComputeInstanceType sets the value of the 'compute_instance_type' attribute
// to the given value.
//
//
func (b *FlavourBuilder) ComputeInstanceType(value string) *FlavourBuilder {
	b.computeInstanceType = &value
	return b
}

// InfraInstanceType sets the value of the 'infra_instance_type' attribute
// to the given value.
//
//
func (b *FlavourBuilder) InfraInstanceType(value string) *FlavourBuilder {
	b.infraInstanceType = &value
	return b
}

// MasterInstanceType sets the value of the 'master_instance_type' attribute
// to the given value.
//
//
func (b *FlavourBuilder) MasterInstanceType(value string) *FlavourBuilder {
	b.masterInstanceType = &value
	return b
}

// Name sets the value of the 'name' attribute
// to the given value.
//
//
func (b *FlavourBuilder) Name(value string) *FlavourBuilder {
	b.name = &value
	return b
}

// Network sets the value of the 'network' attribute
// to the given value.
//
// Network configuration of a cluster.
func (b *FlavourBuilder) Network(value *NetworkBuilder) *FlavourBuilder {
	b.network = value
	return b
}

// Nodes sets the value of the 'nodes' attribute
// to the given value.
//
// Counts of different classes of nodes inside a flavour.
func (b *FlavourBuilder) Nodes(value *FlavourNodesBuilder) *FlavourBuilder {
	b.nodes = value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *FlavourBuilder) Copy(object *Flavour) *FlavourBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	if object.aws != nil {
		b.aws = NewAWSFlavour().Copy(object.aws)
	} else {
		b.aws = nil
	}
	b.computeInstanceType = object.computeInstanceType
	b.infraInstanceType = object.infraInstanceType
	b.masterInstanceType = object.masterInstanceType
	b.name = object.name
	if object.network != nil {
		b.network = NewNetwork().Copy(object.network)
	} else {
		b.network = nil
	}
	if object.nodes != nil {
		b.nodes = NewFlavourNodes().Copy(object.nodes)
	} else {
		b.nodes = nil
	}
	return b
}

// Build creates a 'flavour' object using the configuration stored in the builder.
func (b *FlavourBuilder) Build() (object *Flavour, err error) {
	object = new(Flavour)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	if b.aws != nil {
		object.aws, err = b.aws.Build()
		if err != nil {
			return
		}
	}
	if b.computeInstanceType != nil {
		object.computeInstanceType = b.computeInstanceType
	}
	if b.infraInstanceType != nil {
		object.infraInstanceType = b.infraInstanceType
	}
	if b.masterInstanceType != nil {
		object.masterInstanceType = b.masterInstanceType
	}
	if b.name != nil {
		object.name = b.name
	}
	if b.network != nil {
		object.network, err = b.network.Build()
		if err != nil {
			return
		}
	}
	if b.nodes != nil {
		object.nodes, err = b.nodes.Build()
		if err != nil {
			return
		}
	}
	return
}
