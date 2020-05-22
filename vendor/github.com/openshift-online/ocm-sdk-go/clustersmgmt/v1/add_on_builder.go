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

// AddOnBuilder contains the data and logic needed to build 'add_on' objects.
//
// Representation of an add-on that can be installed in a cluster.
type AddOnBuilder struct {
	id              *string
	href            *string
	link            bool
	description     *string
	docsLink        *string
	enabled         *bool
	icon            *string
	installMode     *AddOnInstallMode
	label           *string
	name            *string
	operatorName    *string
	resourceCost    *float64
	resourceName    *string
	targetNamespace *string
}

// NewAddOn creates a new builder of 'add_on' objects.
func NewAddOn() *AddOnBuilder {
	return new(AddOnBuilder)
}

// ID sets the identifier of the object.
func (b *AddOnBuilder) ID(value string) *AddOnBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *AddOnBuilder) HREF(value string) *AddOnBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *AddOnBuilder) Link(value bool) *AddOnBuilder {
	b.link = value
	return b
}

// Description sets the value of the 'description' attribute to the given value.
//
//
func (b *AddOnBuilder) Description(value string) *AddOnBuilder {
	b.description = &value
	return b
}

// DocsLink sets the value of the 'docs_link' attribute to the given value.
//
//
func (b *AddOnBuilder) DocsLink(value string) *AddOnBuilder {
	b.docsLink = &value
	return b
}

// Enabled sets the value of the 'enabled' attribute to the given value.
//
//
func (b *AddOnBuilder) Enabled(value bool) *AddOnBuilder {
	b.enabled = &value
	return b
}

// Icon sets the value of the 'icon' attribute to the given value.
//
//
func (b *AddOnBuilder) Icon(value string) *AddOnBuilder {
	b.icon = &value
	return b
}

// InstallMode sets the value of the 'install_mode' attribute to the given value.
//
// Representation of an add-on InstallMode field.
func (b *AddOnBuilder) InstallMode(value AddOnInstallMode) *AddOnBuilder {
	b.installMode = &value
	return b
}

// Label sets the value of the 'label' attribute to the given value.
//
//
func (b *AddOnBuilder) Label(value string) *AddOnBuilder {
	b.label = &value
	return b
}

// Name sets the value of the 'name' attribute to the given value.
//
//
func (b *AddOnBuilder) Name(value string) *AddOnBuilder {
	b.name = &value
	return b
}

// OperatorName sets the value of the 'operator_name' attribute to the given value.
//
//
func (b *AddOnBuilder) OperatorName(value string) *AddOnBuilder {
	b.operatorName = &value
	return b
}

// ResourceCost sets the value of the 'resource_cost' attribute to the given value.
//
//
func (b *AddOnBuilder) ResourceCost(value float64) *AddOnBuilder {
	b.resourceCost = &value
	return b
}

// ResourceName sets the value of the 'resource_name' attribute to the given value.
//
//
func (b *AddOnBuilder) ResourceName(value string) *AddOnBuilder {
	b.resourceName = &value
	return b
}

// TargetNamespace sets the value of the 'target_namespace' attribute to the given value.
//
//
func (b *AddOnBuilder) TargetNamespace(value string) *AddOnBuilder {
	b.targetNamespace = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *AddOnBuilder) Copy(object *AddOn) *AddOnBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.description = object.description
	b.docsLink = object.docsLink
	b.enabled = object.enabled
	b.icon = object.icon
	b.installMode = object.installMode
	b.label = object.label
	b.name = object.name
	b.operatorName = object.operatorName
	b.resourceCost = object.resourceCost
	b.resourceName = object.resourceName
	b.targetNamespace = object.targetNamespace
	return b
}

// Build creates a 'add_on' object using the configuration stored in the builder.
func (b *AddOnBuilder) Build() (object *AddOn, err error) {
	object = new(AddOn)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	object.description = b.description
	object.docsLink = b.docsLink
	object.enabled = b.enabled
	object.icon = b.icon
	object.installMode = b.installMode
	object.label = b.label
	object.name = b.name
	object.operatorName = b.operatorName
	object.resourceCost = b.resourceCost
	object.resourceName = b.resourceName
	object.targetNamespace = b.targetNamespace
	return
}
