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
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// roleBindingData is the data structure used internally to marshal and unmarshal
// objects of type 'role_binding'.
type roleBindingData struct {
	Kind           *string           "json:\"kind,omitempty\""
	ID             *string           "json:\"id,omitempty\""
	HREF           *string           "json:\"href,omitempty\""
	Account        *accountData      "json:\"account,omitempty\""
	AccountID      *string           "json:\"account_id,omitempty\""
	ConfigManaged  *bool             "json:\"config_managed,omitempty\""
	Organization   *organizationData "json:\"organization,omitempty\""
	OrganizationID *string           "json:\"organization_id,omitempty\""
	Role           *roleData         "json:\"role,omitempty\""
	RoleID         *string           "json:\"role_id,omitempty\""
	Subscription   *subscriptionData "json:\"subscription,omitempty\""
	SubscriptionID *string           "json:\"subscription_id,omitempty\""
	Type           *string           "json:\"type,omitempty\""
}

// MarshalRoleBinding writes a value of the 'role_binding' to the given target,
// which can be a writer or a JSON encoder.
func MarshalRoleBinding(object *RoleBinding, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'role_binding'
// value to a JSON document.
func (o *RoleBinding) wrap() (data *roleBindingData, err error) {
	if o == nil {
		return
	}
	data = new(roleBindingData)
	data.ID = o.id
	data.HREF = o.href
	data.Kind = new(string)
	if o.link {
		*data.Kind = RoleBindingLinkKind
	} else {
		*data.Kind = RoleBindingKind
	}
	data.Account, err = o.account.wrap()
	if err != nil {
		return
	}
	data.AccountID = o.accountID
	data.ConfigManaged = o.configManaged
	data.Organization, err = o.organization.wrap()
	if err != nil {
		return
	}
	data.OrganizationID = o.organizationID
	data.Role, err = o.role.wrap()
	if err != nil {
		return
	}
	data.RoleID = o.roleID
	data.Subscription, err = o.subscription.wrap()
	if err != nil {
		return
	}
	data.SubscriptionID = o.subscriptionID
	data.Type = o.type_
	return
}

// UnmarshalRoleBinding reads a value of the 'role_binding' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalRoleBinding(source interface{}) (object *RoleBinding, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(roleBindingData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'role_binding' type.
func (d *roleBindingData) unwrap() (object *RoleBinding, err error) {
	if d == nil {
		return
	}
	object = new(RoleBinding)
	object.id = d.ID
	object.href = d.HREF
	if d.Kind != nil {
		object.link = *d.Kind == RoleBindingLinkKind
	}
	object.account, err = d.Account.unwrap()
	if err != nil {
		return
	}
	object.accountID = d.AccountID
	object.configManaged = d.ConfigManaged
	object.organization, err = d.Organization.unwrap()
	if err != nil {
		return
	}
	object.organizationID = d.OrganizationID
	object.role, err = d.Role.unwrap()
	if err != nil {
		return
	}
	object.roleID = d.RoleID
	object.subscription, err = d.Subscription.unwrap()
	if err != nil {
		return
	}
	object.subscriptionID = d.SubscriptionID
	object.type_ = d.Type
	return
}
