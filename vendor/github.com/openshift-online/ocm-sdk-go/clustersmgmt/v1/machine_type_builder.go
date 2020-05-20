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

// MachineTypeBuilder contains the data and logic needed to build 'machine_type' objects.
//
// Machine type.
type MachineTypeBuilder struct {
	id            *string
	href          *string
	link          bool
	cpu           *ValueBuilder
	category      *MachineTypeCategory
	cloudProvider *CloudProviderBuilder
	memory        *ValueBuilder
	name          *string
}

// NewMachineType creates a new builder of 'machine_type' objects.
func NewMachineType() *MachineTypeBuilder {
	return new(MachineTypeBuilder)
}

// ID sets the identifier of the object.
func (b *MachineTypeBuilder) ID(value string) *MachineTypeBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *MachineTypeBuilder) HREF(value string) *MachineTypeBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *MachineTypeBuilder) Link(value bool) *MachineTypeBuilder {
	b.link = value
	return b
}

// CPU sets the value of the 'CPU' attribute to the given value.
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
func (b *MachineTypeBuilder) CPU(value *ValueBuilder) *MachineTypeBuilder {
	b.cpu = value
	return b
}

// Category sets the value of the 'category' attribute to the given value.
//
// Machine type category.
func (b *MachineTypeBuilder) Category(value MachineTypeCategory) *MachineTypeBuilder {
	b.category = &value
	return b
}

// CloudProvider sets the value of the 'cloud_provider' attribute to the given value.
//
// Cloud provider.
func (b *MachineTypeBuilder) CloudProvider(value *CloudProviderBuilder) *MachineTypeBuilder {
	b.cloudProvider = value
	return b
}

// Memory sets the value of the 'memory' attribute to the given value.
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
func (b *MachineTypeBuilder) Memory(value *ValueBuilder) *MachineTypeBuilder {
	b.memory = value
	return b
}

// Name sets the value of the 'name' attribute to the given value.
//
//
func (b *MachineTypeBuilder) Name(value string) *MachineTypeBuilder {
	b.name = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *MachineTypeBuilder) Copy(object *MachineType) *MachineTypeBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	if object.cpu != nil {
		b.cpu = NewValue().Copy(object.cpu)
	} else {
		b.cpu = nil
	}
	b.category = object.category
	if object.cloudProvider != nil {
		b.cloudProvider = NewCloudProvider().Copy(object.cloudProvider)
	} else {
		b.cloudProvider = nil
	}
	if object.memory != nil {
		b.memory = NewValue().Copy(object.memory)
	} else {
		b.memory = nil
	}
	b.name = object.name
	return b
}

// Build creates a 'machine_type' object using the configuration stored in the builder.
func (b *MachineTypeBuilder) Build() (object *MachineType, err error) {
	object = new(MachineType)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	if b.cpu != nil {
		object.cpu, err = b.cpu.Build()
		if err != nil {
			return
		}
	}
	object.category = b.category
	if b.cloudProvider != nil {
		object.cloudProvider, err = b.cloudProvider.Build()
		if err != nil {
			return
		}
	}
	if b.memory != nil {
		object.memory, err = b.memory.Build()
		if err != nil {
			return
		}
	}
	object.name = b.name
	return
}
