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

package v1 // github.com/openshift-online/ocm-sdk-go/authorizations/v1

import (
	"net/http"
	"path"
)

// Client is the client of the 'root' resource.
//
// Root of the tree of resources of the authorization service.
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

// Creates a new request for the method that retrieves the metadata.
func (c *Client) Get() *MetadataRequest {
	return &MetadataRequest{
		transport: c.transport,
		path:      c.path,
		metric:    c.metric,
	}
}

// AccessReview returns the target 'access_review' resource.
//
// Reference to the resource that is used to submit access review requests.
func (c *Client) AccessReview() *AccessReviewClient {
	return NewAccessReviewClient(
		c.transport,
		path.Join(c.path, "access_review"),
		path.Join(c.metric, "access_review"),
	)
}

// ExportControlReview returns the target 'export_control_review' resource.
//
// Reference to the resource that is used to submit export control review requests.
func (c *Client) ExportControlReview() *ExportControlReviewClient {
	return NewExportControlReviewClient(
		c.transport,
		path.Join(c.path, "export_control_review"),
		path.Join(c.metric, "export_control_review"),
	)
}

// ResourceReview returns the target 'resource_review' resource.
//
// Reference to the resource that is used to submit resource review requests.
func (c *Client) ResourceReview() *ResourceReviewClient {
	return NewResourceReviewClient(
		c.transport,
		path.Join(c.path, "resource_review"),
		path.Join(c.metric, "resource_review"),
	)
}

// SelfAccessReview returns the target 'self_access_review' resource.
//
// Reference to the resource that is used to submit self access review requests.
func (c *Client) SelfAccessReview() *SelfAccessReviewClient {
	return NewSelfAccessReviewClient(
		c.transport,
		path.Join(c.path, "self_access_review"),
		path.Join(c.metric, "self_access_review"),
	)
}
