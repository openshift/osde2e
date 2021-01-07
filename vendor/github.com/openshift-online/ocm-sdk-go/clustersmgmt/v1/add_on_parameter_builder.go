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

// AddOnParameterBuilder contains the data and logic needed to build 'add_on_parameter' objects.
//
// Representation of an add-on parameter.
type AddOnParameterBuilder struct {
	id           *string
	href         *string
	link         bool
	addon        *AddOnBuilder
	defaultValue *string
	description  *string
	editable     *bool
	enabled      *bool
	name         *string
	required     *bool
	validation   *string
	valueType    *string
}

// NewAddOnParameter creates a new builder of 'add_on_parameter' objects.
func NewAddOnParameter() *AddOnParameterBuilder {
	return new(AddOnParameterBuilder)
}

// ID sets the identifier of the object.
func (b *AddOnParameterBuilder) ID(value string) *AddOnParameterBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *AddOnParameterBuilder) HREF(value string) *AddOnParameterBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *AddOnParameterBuilder) Link(value bool) *AddOnParameterBuilder {
	b.link = value
	return b
}

// Addon sets the value of the 'addon' attribute to the given value.
//
// Representation of an add-on that can be installed in a cluster.
func (b *AddOnParameterBuilder) Addon(value *AddOnBuilder) *AddOnParameterBuilder {
	b.addon = value
	return b
}

// DefaultValue sets the value of the 'default_value' attribute to the given value.
//
//
func (b *AddOnParameterBuilder) DefaultValue(value string) *AddOnParameterBuilder {
	b.defaultValue = &value
	return b
}

// Description sets the value of the 'description' attribute to the given value.
//
//
func (b *AddOnParameterBuilder) Description(value string) *AddOnParameterBuilder {
	b.description = &value
	return b
}

// Editable sets the value of the 'editable' attribute to the given value.
//
//
func (b *AddOnParameterBuilder) Editable(value bool) *AddOnParameterBuilder {
	b.editable = &value
	return b
}

// Enabled sets the value of the 'enabled' attribute to the given value.
//
//
func (b *AddOnParameterBuilder) Enabled(value bool) *AddOnParameterBuilder {
	b.enabled = &value
	return b
}

// Name sets the value of the 'name' attribute to the given value.
//
//
func (b *AddOnParameterBuilder) Name(value string) *AddOnParameterBuilder {
	b.name = &value
	return b
}

// Required sets the value of the 'required' attribute to the given value.
//
//
func (b *AddOnParameterBuilder) Required(value bool) *AddOnParameterBuilder {
	b.required = &value
	return b
}

// Validation sets the value of the 'validation' attribute to the given value.
//
//
func (b *AddOnParameterBuilder) Validation(value string) *AddOnParameterBuilder {
	b.validation = &value
	return b
}

// ValueType sets the value of the 'value_type' attribute to the given value.
//
//
func (b *AddOnParameterBuilder) ValueType(value string) *AddOnParameterBuilder {
	b.valueType = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *AddOnParameterBuilder) Copy(object *AddOnParameter) *AddOnParameterBuilder {
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
	b.defaultValue = object.defaultValue
	b.description = object.description
	b.editable = object.editable
	b.enabled = object.enabled
	b.name = object.name
	b.required = object.required
	b.validation = object.validation
	b.valueType = object.valueType
	return b
}

// Build creates a 'add_on_parameter' object using the configuration stored in the builder.
func (b *AddOnParameterBuilder) Build() (object *AddOnParameter, err error) {
	object = new(AddOnParameter)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	if b.addon != nil {
		object.addon, err = b.addon.Build()
		if err != nil {
			return
		}
	}
	object.defaultValue = b.defaultValue
	object.description = b.description
	object.editable = b.editable
	object.enabled = b.enabled
	object.name = b.name
	object.required = b.required
	object.validation = b.validation
	object.valueType = b.valueType
	return
}
