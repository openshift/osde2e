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

package v1 // github.com/openshift-online/ocm-sdk-go/authorizations/v1

// SelfAccessReviewResponseBuilder contains the data and logic needed to build 'self_access_review_response' objects.
//
// Representation of an access review response, performed against oneself
type SelfAccessReviewResponseBuilder struct {
	action         *string
	allowed        *bool
	clusterID      *string
	clusterUUID    *string
	organizationID *string
	resourceType   *string
	subscriptionID *string
}

// NewSelfAccessReviewResponse creates a new builder of 'self_access_review_response' objects.
func NewSelfAccessReviewResponse() *SelfAccessReviewResponseBuilder {
	return new(SelfAccessReviewResponseBuilder)
}

// Action sets the value of the 'action' attribute to the given value.
//
//
func (b *SelfAccessReviewResponseBuilder) Action(value string) *SelfAccessReviewResponseBuilder {
	b.action = &value
	return b
}

// Allowed sets the value of the 'allowed' attribute to the given value.
//
//
func (b *SelfAccessReviewResponseBuilder) Allowed(value bool) *SelfAccessReviewResponseBuilder {
	b.allowed = &value
	return b
}

// ClusterID sets the value of the 'cluster_ID' attribute to the given value.
//
//
func (b *SelfAccessReviewResponseBuilder) ClusterID(value string) *SelfAccessReviewResponseBuilder {
	b.clusterID = &value
	return b
}

// ClusterUUID sets the value of the 'cluster_UUID' attribute to the given value.
//
//
func (b *SelfAccessReviewResponseBuilder) ClusterUUID(value string) *SelfAccessReviewResponseBuilder {
	b.clusterUUID = &value
	return b
}

// OrganizationID sets the value of the 'organization_ID' attribute to the given value.
//
//
func (b *SelfAccessReviewResponseBuilder) OrganizationID(value string) *SelfAccessReviewResponseBuilder {
	b.organizationID = &value
	return b
}

// ResourceType sets the value of the 'resource_type' attribute to the given value.
//
//
func (b *SelfAccessReviewResponseBuilder) ResourceType(value string) *SelfAccessReviewResponseBuilder {
	b.resourceType = &value
	return b
}

// SubscriptionID sets the value of the 'subscription_ID' attribute to the given value.
//
//
func (b *SelfAccessReviewResponseBuilder) SubscriptionID(value string) *SelfAccessReviewResponseBuilder {
	b.subscriptionID = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *SelfAccessReviewResponseBuilder) Copy(object *SelfAccessReviewResponse) *SelfAccessReviewResponseBuilder {
	if object == nil {
		return b
	}
	b.action = object.action
	b.allowed = object.allowed
	b.clusterID = object.clusterID
	b.clusterUUID = object.clusterUUID
	b.organizationID = object.organizationID
	b.resourceType = object.resourceType
	b.subscriptionID = object.subscriptionID
	return b
}

// Build creates a 'self_access_review_response' object using the configuration stored in the builder.
func (b *SelfAccessReviewResponseBuilder) Build() (object *SelfAccessReviewResponse, err error) {
	object = new(SelfAccessReviewResponse)
	object.action = b.action
	object.allowed = b.allowed
	object.clusterID = b.clusterID
	object.clusterUUID = b.clusterUUID
	object.organizationID = b.organizationID
	object.resourceType = b.resourceType
	object.subscriptionID = b.subscriptionID
	return
}
