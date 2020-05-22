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

// LogBuilder contains the data and logic needed to build 'log' objects.
//
// Log of the cluster.
type LogBuilder struct {
	id      *string
	href    *string
	link    bool
	content *string
}

// NewLog creates a new builder of 'log' objects.
func NewLog() *LogBuilder {
	return new(LogBuilder)
}

// ID sets the identifier of the object.
func (b *LogBuilder) ID(value string) *LogBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *LogBuilder) HREF(value string) *LogBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *LogBuilder) Link(value bool) *LogBuilder {
	b.link = value
	return b
}

// Content sets the value of the 'content' attribute to the given value.
//
//
func (b *LogBuilder) Content(value string) *LogBuilder {
	b.content = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *LogBuilder) Copy(object *Log) *LogBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.content = object.content
	return b
}

// Build creates a 'log' object using the configuration stored in the builder.
func (b *LogBuilder) Build() (object *Log, err error) {
	object = new(Log)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	object.content = b.content
	return
}
