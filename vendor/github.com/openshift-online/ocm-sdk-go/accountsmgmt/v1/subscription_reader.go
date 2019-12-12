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

package v1 // github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1

import (
	"fmt"
	time "time"

	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// subscriptionData is the data structure used internally to marshal and unmarshal
// objects of type 'subscription'.
type subscriptionData struct {
	Kind               *string                 "json:\"kind,omitempty\""
	ID                 *string                 "json:\"id,omitempty\""
	HREF               *string                 "json:\"href,omitempty\""
	ClusterID          *string                 "json:\"cluster_id,omitempty\""
	Creator            *accountData            "json:\"creator,omitempty\""
	DisplayName        *string                 "json:\"display_name,omitempty\""
	ExternalClusterID  *string                 "json:\"external_cluster_id,omitempty\""
	LastTelemetryDate  *time.Time              "json:\"last_telemetry_date,omitempty\""
	OrganizationID     *string                 "json:\"organization_id,omitempty\""
	Plan               *planData               "json:\"plan,omitempty\""
	RegistryCredential *registryCredentialData "json:\"registry_credential,omitempty\""
}

// MarshalSubscription writes a value of the 'subscription' to the given target,
// which can be a writer or a JSON encoder.
func MarshalSubscription(object *Subscription, target interface{}) error {
	encoder, err := helpers.NewEncoder(target)
	if err != nil {
		return err
	}
	data, err := object.wrap()
	if err != nil {
		return err
	}
	return encoder.Encode(data)
}

// wrap is the method used internally to convert a value of the 'subscription'
// value to a JSON document.
func (o *Subscription) wrap() (data *subscriptionData, err error) {
	if o == nil {
		return
	}
	data = new(subscriptionData)
	data.ID = o.id
	data.HREF = o.href
	data.Kind = new(string)
	if o.link {
		*data.Kind = SubscriptionLinkKind
	} else {
		*data.Kind = SubscriptionKind
	}
	data.ClusterID = o.clusterID
	data.Creator, err = o.creator.wrap()
	if err != nil {
		return
	}
	data.DisplayName = o.displayName
	data.ExternalClusterID = o.externalClusterID
	data.LastTelemetryDate = o.lastTelemetryDate
	data.OrganizationID = o.organizationID
	data.Plan, err = o.plan.wrap()
	if err != nil {
		return
	}
	data.RegistryCredential, err = o.registryCredential.wrap()
	if err != nil {
		return
	}
	return
}

// UnmarshalSubscription reads a value of the 'subscription' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalSubscription(source interface{}) (object *Subscription, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(subscriptionData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'subscription' type.
func (d *subscriptionData) unwrap() (object *Subscription, err error) {
	if d == nil {
		return
	}
	object = new(Subscription)
	object.id = d.ID
	object.href = d.HREF
	if d.Kind != nil {
		switch *d.Kind {
		case SubscriptionKind:
			object.link = false
		case SubscriptionLinkKind:
			object.link = true
		default:
			err = fmt.Errorf(
				"expected kind '%s' or '%s' but got '%s'",
				SubscriptionKind,
				SubscriptionLinkKind,
				*d.Kind,
			)
			return
		}
	}
	object.clusterID = d.ClusterID
	object.creator, err = d.Creator.unwrap()
	if err != nil {
		return
	}
	object.displayName = d.DisplayName
	object.externalClusterID = d.ExternalClusterID
	object.lastTelemetryDate = d.LastTelemetryDate
	object.organizationID = d.OrganizationID
	object.plan, err = d.Plan.unwrap()
	if err != nil {
		return
	}
	object.registryCredential, err = d.RegistryCredential.unwrap()
	if err != nil {
		return
	}
	return
}
