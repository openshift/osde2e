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
// resource using the given transport to send the requests and receive the
// responses.
func NewClient(transport http.RoundTripper, path string, metric string) *Client {
	return &Client{
		transport: transport,
		path:      path,
		metric:    metric,
	}
}

// Creates a new request for the method that retrieves the metadata.
func (c *Client) Get() *MetadataRequest {
	return &MetadataRequest{
		transport: c.transport,
		path:      c.path,
		metric:    c.metric,
	}
}

// AWSInfrastructureAccessRoles returns the target 'AWS_infrastructure_access_roles' resource.
//
// Reference to the resource that manages the collection of AWS
// infrastructure access roles.
func (c *Client) AWSInfrastructureAccessRoles() *AWSInfrastructureAccessRolesClient {
	return NewAWSInfrastructureAccessRolesClient(
		c.transport,
		path.Join(c.path, "aws_infrastructure_access_roles"),
		path.Join(c.metric, "aws_infrastructure_access_roles"),
	)
}

// Addons returns the target 'add_ons' resource.
//
// Reference to the resource that manages the collection of add-ons.
func (c *Client) Addons() *AddOnsClient {
	return NewAddOnsClient(
		c.transport,
		path.Join(c.path, "addons"),
		path.Join(c.metric, "addons"),
	)
}

// CloudProviders returns the target 'cloud_providers' resource.
//
// Reference to the resource that manages the collection of cloud providers.
func (c *Client) CloudProviders() *CloudProvidersClient {
	return NewCloudProvidersClient(
		c.transport,
		path.Join(c.path, "cloud_providers"),
		path.Join(c.metric, "cloud_providers"),
	)
}

// Clusters returns the target 'clusters' resource.
//
// Reference to the resource that manages the collection of clusters.
func (c *Client) Clusters() *ClustersClient {
	return NewClustersClient(
		c.transport,
		path.Join(c.path, "clusters"),
		path.Join(c.metric, "clusters"),
	)
}

// Dashboards returns the target 'dashboards' resource.
//
// Reference to the resource that manages the collection of dashboards.
func (c *Client) Dashboards() *DashboardsClient {
	return NewDashboardsClient(
		c.transport,
		path.Join(c.path, "dashboards"),
		path.Join(c.metric, "dashboards"),
	)
}

// Flavours returns the target 'flavours' resource.
//
// Reference to the service that manages the collection of flavours.
func (c *Client) Flavours() *FlavoursClient {
	return NewFlavoursClient(
		c.transport,
		path.Join(c.path, "flavours"),
		path.Join(c.metric, "flavours"),
	)
}

// MachineTypes returns the target 'machine_types' resource.
//
// Reference to the resource that manage the collection of machine types.
func (c *Client) MachineTypes() *MachineTypesClient {
	return NewMachineTypesClient(
		c.transport,
		path.Join(c.path, "machine_types"),
		path.Join(c.metric, "machine_types"),
	)
}

// Products returns the target 'products' resource.
//
// Reference to the resource that manages the collection of products.
func (c *Client) Products() *ProductsClient {
	return NewProductsClient(
		c.transport,
		path.Join(c.path, "products"),
		path.Join(c.metric, "products"),
	)
}

// Versions returns the target 'versions' resource.
//
// Reference to the resource that manage the collection of versions.
func (c *Client) Versions() *VersionsClient {
	return NewVersionsClient(
		c.transport,
		path.Join(c.path, "versions"),
		path.Join(c.metric, "versions"),
	)
}
