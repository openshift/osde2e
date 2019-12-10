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
	"net/http"
	"path"
)

// Client is the client of the 'root' resource.
//
// Root of the tree of resources of the clusters management service.
type Client struct {
	transport http.RoundTripper
	path      string
	metric    string
}

// NewClient creates a new client for the 'root'
// resource using the given transport to sned the requests and receive the
// responses.
func NewClient(transport http.RoundTripper, path string, metric string) *Client {
	client := new(Client)
	client.transport = transport
	client.path = path
	client.metric = metric
	return client
}

// SKUS returns the target 'SKUS' resource.
//
// Reference to the resource that manages the collection of
// SKUS
func (c *Client) SKUS() *SKUSClient {
	return NewSKUSClient(
		c.transport,
		path.Join(c.path, "skus"),
		path.Join(c.metric, "skus"),
	)
}

// AccessToken returns the target 'access_token' resource.
//
// Reference to the resource that manages generates access tokens.
func (c *Client) AccessToken() *AccessTokenClient {
	return NewAccessTokenClient(
		c.transport,
		path.Join(c.path, "access_token"),
		path.Join(c.metric, "access_token"),
	)
}

// Accounts returns the target 'accounts' resource.
//
// Reference to the resource that manages the collection of accounts.
func (c *Client) Accounts() *AccountsClient {
	return NewAccountsClient(
		c.transport,
		path.Join(c.path, "accounts"),
		path.Join(c.metric, "accounts"),
	)
}

// ClusterAuthorizations returns the target 'cluster_authorizations' resource.
//
// Reference to the resource that manages cluster authorizations.
func (c *Client) ClusterAuthorizations() *ClusterAuthorizationsClient {
	return NewClusterAuthorizationsClient(
		c.transport,
		path.Join(c.path, "cluster_authorizations"),
		path.Join(c.metric, "cluster_authorizations"),
	)
}

// ClusterRegistrations returns the target 'cluster_registrations' resource.
//
// Reference to the resource that manages cluster registrations.
func (c *Client) ClusterRegistrations() *ClusterRegistrationsClient {
	return NewClusterRegistrationsClient(
		c.transport,
		path.Join(c.path, "cluster_registrations"),
		path.Join(c.metric, "cluster_registrations"),
	)
}

// CurrentAccount returns the target 'current_account' resource.
//
// Reference to the resource that manages the current authenticated
// acount.
func (c *Client) CurrentAccount() *CurrentAccountClient {
	return NewCurrentAccountClient(
		c.transport,
		path.Join(c.path, "current_account"),
		path.Join(c.metric, "current_account"),
	)
}

// Organizations returns the target 'organizations' resource.
//
// Reference to the resource that manages the collection of
// organizations.
func (c *Client) Organizations() *OrganizationsClient {
	return NewOrganizationsClient(
		c.transport,
		path.Join(c.path, "organizations"),
		path.Join(c.metric, "organizations"),
	)
}

// Permissions returns the target 'permissions' resource.
//
// Reference to the resource that manages the collection of permissions.
func (c *Client) Permissions() *PermissionsClient {
	return NewPermissionsClient(
		c.transport,
		path.Join(c.path, "permissions"),
		path.Join(c.metric, "permissions"),
	)
}

// Registries returns the target 'registries' resource.
//
// Reference to the resource that manages the collection of registries.
func (c *Client) Registries() *RegistriesClient {
	return NewRegistriesClient(
		c.transport,
		path.Join(c.path, "registries"),
		path.Join(c.metric, "registries"),
	)
}

// RegistryCredentials returns the target 'registry_credentials' resource.
//
// Reference to the resource that manages the collection of registry
// credentials.
func (c *Client) RegistryCredentials() *RegistryCredentialsClient {
	return NewRegistryCredentialsClient(
		c.transport,
		path.Join(c.path, "registry_credentials"),
		path.Join(c.metric, "registry_credentials"),
	)
}

// RoleBindings returns the target 'role_bindings' resource.
//
// Reference to the resource that manages the collection of role
// bindings.
func (c *Client) RoleBindings() *RoleBindingsClient {
	return NewRoleBindingsClient(
		c.transport,
		path.Join(c.path, "role_bindings"),
		path.Join(c.metric, "role_bindings"),
	)
}

// Roles returns the target 'roles' resource.
//
// Reference to the resource that manages the collection of roles.
func (c *Client) Roles() *RolesClient {
	return NewRolesClient(
		c.transport,
		path.Join(c.path, "roles"),
		path.Join(c.metric, "roles"),
	)
}

// Subscriptions returns the target 'subscriptions' resource.
//
// Reference to the resource that manages the collection of
// subscriptions.
func (c *Client) Subscriptions() *SubscriptionsClient {
	return NewSubscriptionsClient(
		c.transport,
		path.Join(c.path, "subscriptions"),
		path.Join(c.metric, "subscriptions"),
	)
}
