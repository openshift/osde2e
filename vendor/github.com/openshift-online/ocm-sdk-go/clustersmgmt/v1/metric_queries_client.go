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

package v1 // github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1

import (
	"net/http"
	"path"
)

// MetricQueriesClient is the client of the 'metric_queries' resource.
//
// Manages metric queries for a cluster.
type MetricQueriesClient struct {
	transport http.RoundTripper
	path      string
	metric    string
}

// NewMetricQueriesClient creates a new client for the 'metric_queries'
// resource using the given transport to sned the requests and receive the
// responses.
func NewMetricQueriesClient(transport http.RoundTripper, path string, metric string) *MetricQueriesClient {
	client := new(MetricQueriesClient)
	client.transport = transport
	client.path = path
	client.metric = metric
	return client
}

// CPUTotalByNodeRolesOS returns the target 'CPU_total_by_node_roles_OS_metric_query' resource.
//
// Reference to the resource that retrieves the total cpu
// capacity in the cluster by node role and operating system.
func (c *MetricQueriesClient) CPUTotalByNodeRolesOS() *CPUTotalByNodeRolesOSMetricQueryClient {
	return NewCPUTotalByNodeRolesOSMetricQueryClient(
		c.transport,
		path.Join(c.path, "cpu_total_by_node_roles_os"),
		path.Join(c.metric, "cpu_total_by_node_roles_os"),
	)
}

// SocketTotalByNodeRolesOS returns the target 'socket_total_by_node_roles_OS_metric_query' resource.
//
// Reference to the resource that retrieves the total socket
// capacity in the cluster by node role and operating system.
func (c *MetricQueriesClient) SocketTotalByNodeRolesOS() *SocketTotalByNodeRolesOSMetricQueryClient {
	return NewSocketTotalByNodeRolesOSMetricQueryClient(
		c.transport,
		path.Join(c.path, "socket_total_by_node_roles_os"),
		path.Join(c.metric, "socket_total_by_node_roles_os"),
	)
}
