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

// SupportCaseRequestBuilder contains the data and logic needed to build 'support_case_request' objects.
//
//
type SupportCaseRequestBuilder struct {
	id             *string
	href           *string
	link           bool
	clusterId      *string
	clusterUuid    *string
	description    *string
	eventStreamId  *string
	severity       *string
	subscriptionId *string
	summary        *string
}

// NewSupportCaseRequest creates a new builder of 'support_case_request' objects.
func NewSupportCaseRequest() *SupportCaseRequestBuilder {
	return new(SupportCaseRequestBuilder)
}

// ID sets the identifier of the object.
func (b *SupportCaseRequestBuilder) ID(value string) *SupportCaseRequestBuilder {
	b.id = &value
	return b
}

// HREF sets the link to the object.
func (b *SupportCaseRequestBuilder) HREF(value string) *SupportCaseRequestBuilder {
	b.href = &value
	return b
}

// Link sets the flag that indicates if this is a link.
func (b *SupportCaseRequestBuilder) Link(value bool) *SupportCaseRequestBuilder {
	b.link = value
	return b
}

// ClusterId sets the value of the 'cluster_id' attribute to the given value.
//
//
func (b *SupportCaseRequestBuilder) ClusterId(value string) *SupportCaseRequestBuilder {
	b.clusterId = &value
	return b
}

// ClusterUuid sets the value of the 'cluster_uuid' attribute to the given value.
//
//
func (b *SupportCaseRequestBuilder) ClusterUuid(value string) *SupportCaseRequestBuilder {
	b.clusterUuid = &value
	return b
}

// Description sets the value of the 'description' attribute to the given value.
//
//
func (b *SupportCaseRequestBuilder) Description(value string) *SupportCaseRequestBuilder {
	b.description = &value
	return b
}

// EventStreamId sets the value of the 'event_stream_id' attribute to the given value.
//
//
func (b *SupportCaseRequestBuilder) EventStreamId(value string) *SupportCaseRequestBuilder {
	b.eventStreamId = &value
	return b
}

// Severity sets the value of the 'severity' attribute to the given value.
//
//
func (b *SupportCaseRequestBuilder) Severity(value string) *SupportCaseRequestBuilder {
	b.severity = &value
	return b
}

// SubscriptionId sets the value of the 'subscription_id' attribute to the given value.
//
//
func (b *SupportCaseRequestBuilder) SubscriptionId(value string) *SupportCaseRequestBuilder {
	b.subscriptionId = &value
	return b
}

// Summary sets the value of the 'summary' attribute to the given value.
//
//
func (b *SupportCaseRequestBuilder) Summary(value string) *SupportCaseRequestBuilder {
	b.summary = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *SupportCaseRequestBuilder) Copy(object *SupportCaseRequest) *SupportCaseRequestBuilder {
	if object == nil {
		return b
	}
	b.id = object.id
	b.href = object.href
	b.link = object.link
	b.clusterId = object.clusterId
	b.clusterUuid = object.clusterUuid
	b.description = object.description
	b.eventStreamId = object.eventStreamId
	b.severity = object.severity
	b.subscriptionId = object.subscriptionId
	b.summary = object.summary
	return b
}

// Build creates a 'support_case_request' object using the configuration stored in the builder.
func (b *SupportCaseRequestBuilder) Build() (object *SupportCaseRequest, err error) {
	object = new(SupportCaseRequest)
	object.id = b.id
	object.href = b.href
	object.link = b.link
	object.clusterId = b.clusterId
	object.clusterUuid = b.clusterUuid
	object.description = b.description
	object.eventStreamId = b.eventStreamId
	object.severity = b.severity
	object.subscriptionId = b.subscriptionId
	object.summary = b.summary
	return
}
