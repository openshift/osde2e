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

// ClusterAuthorizationRequestBuilder contains the data and logic needed to build 'cluster_authorization_request' objects.
//
//
type ClusterAuthorizationRequestBuilder struct {
	byoc              *bool
	accountUsername   *string
	availabilityZone  *string
	clusterID         *string
	disconnected      *bool
	displayName       *string
	externalClusterID *string
	managed           *bool
	reserve           *bool
	resources         []*ReservedResourceBuilder
}

// NewClusterAuthorizationRequest creates a new builder of 'cluster_authorization_request' objects.
func NewClusterAuthorizationRequest() *ClusterAuthorizationRequestBuilder {
	return new(ClusterAuthorizationRequestBuilder)
}

// BYOC sets the value of the 'BYOC' attribute to the given value.
//
//
func (b *ClusterAuthorizationRequestBuilder) BYOC(value bool) *ClusterAuthorizationRequestBuilder {
	b.byoc = &value
	return b
}

// AccountUsername sets the value of the 'account_username' attribute to the given value.
//
//
func (b *ClusterAuthorizationRequestBuilder) AccountUsername(value string) *ClusterAuthorizationRequestBuilder {
	b.accountUsername = &value
	return b
}

// AvailabilityZone sets the value of the 'availability_zone' attribute to the given value.
//
//
func (b *ClusterAuthorizationRequestBuilder) AvailabilityZone(value string) *ClusterAuthorizationRequestBuilder {
	b.availabilityZone = &value
	return b
}

// ClusterID sets the value of the 'cluster_ID' attribute to the given value.
//
//
func (b *ClusterAuthorizationRequestBuilder) ClusterID(value string) *ClusterAuthorizationRequestBuilder {
	b.clusterID = &value
	return b
}

// Disconnected sets the value of the 'disconnected' attribute to the given value.
//
//
func (b *ClusterAuthorizationRequestBuilder) Disconnected(value bool) *ClusterAuthorizationRequestBuilder {
	b.disconnected = &value
	return b
}

// DisplayName sets the value of the 'display_name' attribute to the given value.
//
//
func (b *ClusterAuthorizationRequestBuilder) DisplayName(value string) *ClusterAuthorizationRequestBuilder {
	b.displayName = &value
	return b
}

// ExternalClusterID sets the value of the 'external_cluster_ID' attribute to the given value.
//
//
func (b *ClusterAuthorizationRequestBuilder) ExternalClusterID(value string) *ClusterAuthorizationRequestBuilder {
	b.externalClusterID = &value
	return b
}

// Managed sets the value of the 'managed' attribute to the given value.
//
//
func (b *ClusterAuthorizationRequestBuilder) Managed(value bool) *ClusterAuthorizationRequestBuilder {
	b.managed = &value
	return b
}

// Reserve sets the value of the 'reserve' attribute to the given value.
//
//
func (b *ClusterAuthorizationRequestBuilder) Reserve(value bool) *ClusterAuthorizationRequestBuilder {
	b.reserve = &value
	return b
}

// Resources sets the value of the 'resources' attribute to the given values.
//
//
func (b *ClusterAuthorizationRequestBuilder) Resources(values ...*ReservedResourceBuilder) *ClusterAuthorizationRequestBuilder {
	b.resources = make([]*ReservedResourceBuilder, len(values))
	copy(b.resources, values)
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *ClusterAuthorizationRequestBuilder) Copy(object *ClusterAuthorizationRequest) *ClusterAuthorizationRequestBuilder {
	if object == nil {
		return b
	}
	b.byoc = object.byoc
	b.accountUsername = object.accountUsername
	b.availabilityZone = object.availabilityZone
	b.clusterID = object.clusterID
	b.disconnected = object.disconnected
	b.displayName = object.displayName
	b.externalClusterID = object.externalClusterID
	b.managed = object.managed
	b.reserve = object.reserve
	if object.resources != nil {
		b.resources = make([]*ReservedResourceBuilder, len(object.resources))
		for i, v := range object.resources {
			b.resources[i] = NewReservedResource().Copy(v)
		}
	} else {
		b.resources = nil
	}
	return b
}

// Build creates a 'cluster_authorization_request' object using the configuration stored in the builder.
func (b *ClusterAuthorizationRequestBuilder) Build() (object *ClusterAuthorizationRequest, err error) {
	object = new(ClusterAuthorizationRequest)
	object.byoc = b.byoc
	object.accountUsername = b.accountUsername
	object.availabilityZone = b.availabilityZone
	object.clusterID = b.clusterID
	object.disconnected = b.disconnected
	object.displayName = b.displayName
	object.externalClusterID = b.externalClusterID
	object.managed = b.managed
	object.reserve = b.reserve
	if b.resources != nil {
		object.resources = make([]*ReservedResource, len(b.resources))
		for i, v := range b.resources {
			object.resources[i], err = v.Build()
			if err != nil {
				return
			}
		}
	}
	return
}
