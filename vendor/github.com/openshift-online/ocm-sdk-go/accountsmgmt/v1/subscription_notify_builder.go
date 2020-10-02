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

package v1 // github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1

// SubscriptionNotifyBuilder contains the data and logic needed to build 'subscription_notify' objects.
//
// This struct is a request to send a templated email to a user related to this
// subscription.
type SubscriptionNotifyBuilder struct {
	bccAddress         *string
	clusterID          *string
	clusterUUID        *string
	subject            *string
	subscriptionID     *string
	templateName       *string
	templateParameters []*TemplateParameterBuilder
}

// NewSubscriptionNotify creates a new builder of 'subscription_notify' objects.
func NewSubscriptionNotify() *SubscriptionNotifyBuilder {
	return new(SubscriptionNotifyBuilder)
}

// BccAddress sets the value of the 'bcc_address' attribute to the given value.
//
//
func (b *SubscriptionNotifyBuilder) BccAddress(value string) *SubscriptionNotifyBuilder {
	b.bccAddress = &value
	return b
}

// ClusterID sets the value of the 'cluster_ID' attribute to the given value.
//
//
func (b *SubscriptionNotifyBuilder) ClusterID(value string) *SubscriptionNotifyBuilder {
	b.clusterID = &value
	return b
}

// ClusterUUID sets the value of the 'cluster_UUID' attribute to the given value.
//
//
func (b *SubscriptionNotifyBuilder) ClusterUUID(value string) *SubscriptionNotifyBuilder {
	b.clusterUUID = &value
	return b
}

// Subject sets the value of the 'subject' attribute to the given value.
//
//
func (b *SubscriptionNotifyBuilder) Subject(value string) *SubscriptionNotifyBuilder {
	b.subject = &value
	return b
}

// SubscriptionID sets the value of the 'subscription_ID' attribute to the given value.
//
//
func (b *SubscriptionNotifyBuilder) SubscriptionID(value string) *SubscriptionNotifyBuilder {
	b.subscriptionID = &value
	return b
}

// TemplateName sets the value of the 'template_name' attribute to the given value.
//
//
func (b *SubscriptionNotifyBuilder) TemplateName(value string) *SubscriptionNotifyBuilder {
	b.templateName = &value
	return b
}

// TemplateParameters sets the value of the 'template_parameters' attribute to the given values.
//
//
func (b *SubscriptionNotifyBuilder) TemplateParameters(values ...*TemplateParameterBuilder) *SubscriptionNotifyBuilder {
	b.templateParameters = make([]*TemplateParameterBuilder, len(values))
	copy(b.templateParameters, values)
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *SubscriptionNotifyBuilder) Copy(object *SubscriptionNotify) *SubscriptionNotifyBuilder {
	if object == nil {
		return b
	}
	b.bccAddress = object.bccAddress
	b.clusterID = object.clusterID
	b.clusterUUID = object.clusterUUID
	b.subject = object.subject
	b.subscriptionID = object.subscriptionID
	b.templateName = object.templateName
	if object.templateParameters != nil {
		b.templateParameters = make([]*TemplateParameterBuilder, len(object.templateParameters))
		for i, v := range object.templateParameters {
			b.templateParameters[i] = NewTemplateParameter().Copy(v)
		}
	} else {
		b.templateParameters = nil
	}
	return b
}

// Build creates a 'subscription_notify' object using the configuration stored in the builder.
func (b *SubscriptionNotifyBuilder) Build() (object *SubscriptionNotify, err error) {
	object = new(SubscriptionNotify)
	object.bccAddress = b.bccAddress
	object.clusterID = b.clusterID
	object.clusterUUID = b.clusterUUID
	object.subject = b.subject
	object.subscriptionID = b.subscriptionID
	object.templateName = b.templateName
	if b.templateParameters != nil {
		object.templateParameters = make([]*TemplateParameter, len(b.templateParameters))
		for i, v := range b.templateParameters {
			object.templateParameters[i], err = v.Build()
			if err != nil {
				return
			}
		}
	}
	return
}
