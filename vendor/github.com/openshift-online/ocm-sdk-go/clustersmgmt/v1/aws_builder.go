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

// AWSBuilder contains the data and logic needed to build 'AWS' objects.
//
// _Amazon Web Services_ specific settings of a cluster.
type AWSBuilder struct {
	accessKeyID     *string
	accountID       *string
	secretAccessKey *string
}

// NewAWS creates a new builder of 'AWS' objects.
func NewAWS() *AWSBuilder {
	return new(AWSBuilder)
}

// AccessKeyID sets the value of the 'access_key_ID' attribute to the given value.
//
//
func (b *AWSBuilder) AccessKeyID(value string) *AWSBuilder {
	b.accessKeyID = &value
	return b
}

// AccountID sets the value of the 'account_ID' attribute to the given value.
//
//
func (b *AWSBuilder) AccountID(value string) *AWSBuilder {
	b.accountID = &value
	return b
}

// SecretAccessKey sets the value of the 'secret_access_key' attribute to the given value.
//
//
func (b *AWSBuilder) SecretAccessKey(value string) *AWSBuilder {
	b.secretAccessKey = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *AWSBuilder) Copy(object *AWS) *AWSBuilder {
	if object == nil {
		return b
	}
	b.accessKeyID = object.accessKeyID
	b.accountID = object.accountID
	b.secretAccessKey = object.secretAccessKey
	return b
}

// Build creates a 'AWS' object using the configuration stored in the builder.
func (b *AWSBuilder) Build() (object *AWS, err error) {
	object = new(AWS)
	object.accessKeyID = b.accessKeyID
	object.accountID = b.accountID
	object.secretAccessKey = b.secretAccessKey
	return
}
