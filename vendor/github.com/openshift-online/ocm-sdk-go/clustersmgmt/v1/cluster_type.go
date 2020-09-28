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
	time "time"
)

// ClusterKind is the name of the type used to represent objects
// of type 'cluster'.
const ClusterKind = "Cluster"

// ClusterLinkKind is the name of the type used to represent links
// to objects of type 'cluster'.
const ClusterLinkKind = "ClusterLink"

// ClusterNilKind is the name of the type used to nil references
// to objects of type 'cluster'.
const ClusterNilKind = "ClusterNil"

// Cluster represents the values of the 'cluster' type.
//
// Definition of an _OpenShift_ cluster.
//
// The `cloud_provider` attribute is a reference to the cloud provider. When a
// cluster is retrieved it will be a link to the cloud provider, containing only
// the kind, id and href attributes:
//
// [source,json]
// ----
// {
//   "cloud_provider": {
//     "kind": "CloudProviderLink",
//     "id": "123",
//     "href": "/api/clusters_mgmt/v1/cloud_providers/123"
//   }
// }
// ----
//
// When a cluster is created this is optional, and if used it should contain the
// identifier of the cloud provider to use:
//
// [source,json]
// ----
// {
//   "cloud_provider": {
//     "id": "123",
//   }
// }
// ----
//
// If not included, then the cluster will be created using the default cloud
// provider, which is currently Amazon Web Services.
//
// The region attribute is mandatory when a cluster is created.
//
// The `aws.access_key_id`, `aws.secret_access_key` and `dns.base_domain`
// attributes are mandatory when creation a cluster with your own Amazon Web
// Services account.
type Cluster struct {
	id                                *string
	href                              *string
	link                              bool
	api                               *ClusterAPI
	aws                               *AWS
	awsInfrastructureAccessRoleGrants *AWSInfrastructureAccessRoleGrantList
	byoc                              *bool
	ccs                               *CCS
	dns                               *DNS
	dnsReady                          *bool
	addons                            *AddOnInstallationList
	cloudProvider                     *CloudProvider
	clusterAdminEnabled               *bool
	console                           *ClusterConsole
	creationTimestamp                 *time.Time
	displayName                       *string
	expirationTimestamp               *time.Time
	externalID                        *string
	externalConfiguration             *ExternalConfiguration
	flavour                           *Flavour
	groups                            *GroupList
	healthState                       *ClusterHealthState
	identityProviders                 *IdentityProviderList
	ingresses                         *IngressList
	loadBalancerQuota                 *int
	managed                           *bool
	metrics                           *ClusterMetrics
	multiAZ                           *bool
	name                              *string
	network                           *Network
	nodes                             *ClusterNodes
	openshiftVersion                  *string
	product                           *Product
	properties                        map[string]string
	provisionShard                    *ProvisionShard
	region                            *CloudRegion
	state                             *ClusterState
	status                            *ClusterStatus
	storageQuota                      *Value
	subscription                      *Subscription
	version                           *Version
}

// Kind returns the name of the type of the object.
func (o *Cluster) Kind() string {
	if o == nil {
		return ClusterNilKind
	}
	if o.link {
		return ClusterLinkKind
	}
	return ClusterKind
}

// ID returns the identifier of the object.
func (o *Cluster) ID() string {
	if o != nil && o.id != nil {
		return *o.id
	}
	return ""
}

// GetID returns the identifier of the object and a flag indicating if the
// identifier has a value.
func (o *Cluster) GetID() (value string, ok bool) {
	ok = o != nil && o.id != nil
	if ok {
		value = *o.id
	}
	return
}

// Link returns true iif this is a link.
func (o *Cluster) Link() bool {
	return o != nil && o.link
}

// HREF returns the link to the object.
func (o *Cluster) HREF() string {
	if o != nil && o.href != nil {
		return *o.href
	}
	return ""
}

// GetHREF returns the link of the object and a flag indicating if the
// link has a value.
func (o *Cluster) GetHREF() (value string, ok bool) {
	ok = o != nil && o.href != nil
	if ok {
		value = *o.href
	}
	return
}

// Empty returns true if the object is empty, i.e. no attribute has a value.
func (o *Cluster) Empty() bool {
	return o == nil || (o.id == nil &&
		o.awsInfrastructureAccessRoleGrants.Len() == 0 &&
		o.byoc == nil &&
		o.dnsReady == nil &&
		o.addons.Len() == 0 &&
		o.clusterAdminEnabled == nil &&
		o.creationTimestamp == nil &&
		o.displayName == nil &&
		o.expirationTimestamp == nil &&
		o.externalID == nil &&
		o.groups.Len() == 0 &&
		o.healthState == nil &&
		o.identityProviders.Len() == 0 &&
		o.ingresses.Len() == 0 &&
		o.loadBalancerQuota == nil &&
		o.managed == nil &&
		o.multiAZ == nil &&
		o.name == nil &&
		o.openshiftVersion == nil &&
		len(o.properties) == 0 &&
		o.state == nil &&
		true)
}

// API returns the value of the 'API' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Information about the API of the cluster.
func (o *Cluster) API() *ClusterAPI {
	if o == nil {
		return nil
	}
	return o.api
}

// GetAPI returns the value of the 'API' attribute and
// a flag indicating if the attribute has a value.
//
// Information about the API of the cluster.
func (o *Cluster) GetAPI() (value *ClusterAPI, ok bool) {
	ok = o != nil && o.api != nil
	if ok {
		value = o.api
	}
	return
}

// AWS returns the value of the 'AWS' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Amazon Web Services settings of the cluster.
func (o *Cluster) AWS() *AWS {
	if o == nil {
		return nil
	}
	return o.aws
}

// GetAWS returns the value of the 'AWS' attribute and
// a flag indicating if the attribute has a value.
//
// Amazon Web Services settings of the cluster.
func (o *Cluster) GetAWS() (value *AWS, ok bool) {
	ok = o != nil && o.aws != nil
	if ok {
		value = o.aws
	}
	return
}

// AWSInfrastructureAccessRoleGrants returns the value of the 'AWS_infrastructure_access_role_grants' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// List of AWS infrastructure access role grants on this cluster.
func (o *Cluster) AWSInfrastructureAccessRoleGrants() *AWSInfrastructureAccessRoleGrantList {
	if o == nil {
		return nil
	}
	return o.awsInfrastructureAccessRoleGrants
}

// GetAWSInfrastructureAccessRoleGrants returns the value of the 'AWS_infrastructure_access_role_grants' attribute and
// a flag indicating if the attribute has a value.
//
// List of AWS infrastructure access role grants on this cluster.
func (o *Cluster) GetAWSInfrastructureAccessRoleGrants() (value *AWSInfrastructureAccessRoleGrantList, ok bool) {
	ok = o != nil && o.awsInfrastructureAccessRoleGrants != nil
	if ok {
		value = o.awsInfrastructureAccessRoleGrants
	}
	return
}

// BYOC returns the value of the 'BYOC' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Flag indicating if the cluster is BYOC (customer cloud subscription).
func (o *Cluster) BYOC() bool {
	if o != nil && o.byoc != nil {
		return *o.byoc
	}
	return false
}

// GetBYOC returns the value of the 'BYOC' attribute and
// a flag indicating if the attribute has a value.
//
// Flag indicating if the cluster is BYOC (customer cloud subscription).
func (o *Cluster) GetBYOC() (value bool, ok bool) {
	ok = o != nil && o.byoc != nil
	if ok {
		value = *o.byoc
	}
	return
}

// CCS returns the value of the 'CCS' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Contains configuration of a Customer Cloud Subscription cluster.
func (o *Cluster) CCS() *CCS {
	if o == nil {
		return nil
	}
	return o.ccs
}

// GetCCS returns the value of the 'CCS' attribute and
// a flag indicating if the attribute has a value.
//
// Contains configuration of a Customer Cloud Subscription cluster.
func (o *Cluster) GetCCS() (value *CCS, ok bool) {
	ok = o != nil && o.ccs != nil
	if ok {
		value = o.ccs
	}
	return
}

// DNS returns the value of the 'DNS' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// DNS settings of the cluster.
func (o *Cluster) DNS() *DNS {
	if o == nil {
		return nil
	}
	return o.dns
}

// GetDNS returns the value of the 'DNS' attribute and
// a flag indicating if the attribute has a value.
//
// DNS settings of the cluster.
func (o *Cluster) GetDNS() (value *DNS, ok bool) {
	ok = o != nil && o.dns != nil
	if ok {
		value = o.dns
	}
	return
}

// DNSReady returns the value of the 'DNS_ready' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Provisioner DNS Ready Status
func (o *Cluster) DNSReady() bool {
	if o != nil && o.dnsReady != nil {
		return *o.dnsReady
	}
	return false
}

// GetDNSReady returns the value of the 'DNS_ready' attribute and
// a flag indicating if the attribute has a value.
//
// Provisioner DNS Ready Status
func (o *Cluster) GetDNSReady() (value bool, ok bool) {
	ok = o != nil && o.dnsReady != nil
	if ok {
		value = *o.dnsReady
	}
	return
}

// Addons returns the value of the 'addons' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// List of add-ons on this cluster.
func (o *Cluster) Addons() *AddOnInstallationList {
	if o == nil {
		return nil
	}
	return o.addons
}

// GetAddons returns the value of the 'addons' attribute and
// a flag indicating if the attribute has a value.
//
// List of add-ons on this cluster.
func (o *Cluster) GetAddons() (value *AddOnInstallationList, ok bool) {
	ok = o != nil && o.addons != nil
	if ok {
		value = o.addons
	}
	return
}

// CloudProvider returns the value of the 'cloud_provider' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Link to the cloud provider where the cluster is installed.
func (o *Cluster) CloudProvider() *CloudProvider {
	if o == nil {
		return nil
	}
	return o.cloudProvider
}

// GetCloudProvider returns the value of the 'cloud_provider' attribute and
// a flag indicating if the attribute has a value.
//
// Link to the cloud provider where the cluster is installed.
func (o *Cluster) GetCloudProvider() (value *CloudProvider, ok bool) {
	ok = o != nil && o.cloudProvider != nil
	if ok {
		value = o.cloudProvider
	}
	return
}

// ClusterAdminEnabled returns the value of the 'cluster_admin_enabled' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Flag indicating if cluster-admin access is enabled.
func (o *Cluster) ClusterAdminEnabled() bool {
	if o != nil && o.clusterAdminEnabled != nil {
		return *o.clusterAdminEnabled
	}
	return false
}

// GetClusterAdminEnabled returns the value of the 'cluster_admin_enabled' attribute and
// a flag indicating if the attribute has a value.
//
// Flag indicating if cluster-admin access is enabled.
func (o *Cluster) GetClusterAdminEnabled() (value bool, ok bool) {
	ok = o != nil && o.clusterAdminEnabled != nil
	if ok {
		value = *o.clusterAdminEnabled
	}
	return
}

// Console returns the value of the 'console' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Information about the console of the cluster.
func (o *Cluster) Console() *ClusterConsole {
	if o == nil {
		return nil
	}
	return o.console
}

// GetConsole returns the value of the 'console' attribute and
// a flag indicating if the attribute has a value.
//
// Information about the console of the cluster.
func (o *Cluster) GetConsole() (value *ClusterConsole, ok bool) {
	ok = o != nil && o.console != nil
	if ok {
		value = o.console
	}
	return
}

// CreationTimestamp returns the value of the 'creation_timestamp' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Date and time when the cluster was initially created, using the
// format defined in https://www.ietf.org/rfc/rfc3339.txt[RC3339].
func (o *Cluster) CreationTimestamp() time.Time {
	if o != nil && o.creationTimestamp != nil {
		return *o.creationTimestamp
	}
	return time.Time{}
}

// GetCreationTimestamp returns the value of the 'creation_timestamp' attribute and
// a flag indicating if the attribute has a value.
//
// Date and time when the cluster was initially created, using the
// format defined in https://www.ietf.org/rfc/rfc3339.txt[RC3339].
func (o *Cluster) GetCreationTimestamp() (value time.Time, ok bool) {
	ok = o != nil && o.creationTimestamp != nil
	if ok {
		value = *o.creationTimestamp
	}
	return
}

// DisplayName returns the value of the 'display_name' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Name of the cluster for display purposes. It can contain any
// characters, including spaces.
func (o *Cluster) DisplayName() string {
	if o != nil && o.displayName != nil {
		return *o.displayName
	}
	return ""
}

// GetDisplayName returns the value of the 'display_name' attribute and
// a flag indicating if the attribute has a value.
//
// Name of the cluster for display purposes. It can contain any
// characters, including spaces.
func (o *Cluster) GetDisplayName() (value string, ok bool) {
	ok = o != nil && o.displayName != nil
	if ok {
		value = *o.displayName
	}
	return
}

// ExpirationTimestamp returns the value of the 'expiration_timestamp' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Date and time when the cluster will be automatically deleted, using the format defined in
// https://www.ietf.org/rfc/rfc3339.txt[RFC3339]. If no timestamp is provided, the cluster
// will never expire.
//
// This option is unsupported.
func (o *Cluster) ExpirationTimestamp() time.Time {
	if o != nil && o.expirationTimestamp != nil {
		return *o.expirationTimestamp
	}
	return time.Time{}
}

// GetExpirationTimestamp returns the value of the 'expiration_timestamp' attribute and
// a flag indicating if the attribute has a value.
//
// Date and time when the cluster will be automatically deleted, using the format defined in
// https://www.ietf.org/rfc/rfc3339.txt[RFC3339]. If no timestamp is provided, the cluster
// will never expire.
//
// This option is unsupported.
func (o *Cluster) GetExpirationTimestamp() (value time.Time, ok bool) {
	ok = o != nil && o.expirationTimestamp != nil
	if ok {
		value = *o.expirationTimestamp
	}
	return
}

// ExternalID returns the value of the 'external_ID' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// External identifier of the cluster, generated by the installer.
func (o *Cluster) ExternalID() string {
	if o != nil && o.externalID != nil {
		return *o.externalID
	}
	return ""
}

// GetExternalID returns the value of the 'external_ID' attribute and
// a flag indicating if the attribute has a value.
//
// External identifier of the cluster, generated by the installer.
func (o *Cluster) GetExternalID() (value string, ok bool) {
	ok = o != nil && o.externalID != nil
	if ok {
		value = *o.externalID
	}
	return
}

// ExternalConfiguration returns the value of the 'external_configuration' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// ExternalConfiguration shows external configuration on the cluster.
func (o *Cluster) ExternalConfiguration() *ExternalConfiguration {
	if o == nil {
		return nil
	}
	return o.externalConfiguration
}

// GetExternalConfiguration returns the value of the 'external_configuration' attribute and
// a flag indicating if the attribute has a value.
//
// ExternalConfiguration shows external configuration on the cluster.
func (o *Cluster) GetExternalConfiguration() (value *ExternalConfiguration, ok bool) {
	ok = o != nil && o.externalConfiguration != nil
	if ok {
		value = o.externalConfiguration
	}
	return
}

// Flavour returns the value of the 'flavour' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Link to the _flavour_ that was used to create the cluster.
func (o *Cluster) Flavour() *Flavour {
	if o == nil {
		return nil
	}
	return o.flavour
}

// GetFlavour returns the value of the 'flavour' attribute and
// a flag indicating if the attribute has a value.
//
// Link to the _flavour_ that was used to create the cluster.
func (o *Cluster) GetFlavour() (value *Flavour, ok bool) {
	ok = o != nil && o.flavour != nil
	if ok {
		value = o.flavour
	}
	return
}

// Groups returns the value of the 'groups' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Link to the collection of groups of user of the cluster.
func (o *Cluster) Groups() *GroupList {
	if o == nil {
		return nil
	}
	return o.groups
}

// GetGroups returns the value of the 'groups' attribute and
// a flag indicating if the attribute has a value.
//
// Link to the collection of groups of user of the cluster.
func (o *Cluster) GetGroups() (value *GroupList, ok bool) {
	ok = o != nil && o.groups != nil
	if ok {
		value = o.groups
	}
	return
}

// HealthState returns the value of the 'health_state' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// HealthState indicates the overall health state of the cluster.
func (o *Cluster) HealthState() ClusterHealthState {
	if o != nil && o.healthState != nil {
		return *o.healthState
	}
	return ClusterHealthState("")
}

// GetHealthState returns the value of the 'health_state' attribute and
// a flag indicating if the attribute has a value.
//
// HealthState indicates the overall health state of the cluster.
func (o *Cluster) GetHealthState() (value ClusterHealthState, ok bool) {
	ok = o != nil && o.healthState != nil
	if ok {
		value = *o.healthState
	}
	return
}

// IdentityProviders returns the value of the 'identity_providers' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Link to the collection of identity providers of the cluster.
func (o *Cluster) IdentityProviders() *IdentityProviderList {
	if o == nil {
		return nil
	}
	return o.identityProviders
}

// GetIdentityProviders returns the value of the 'identity_providers' attribute and
// a flag indicating if the attribute has a value.
//
// Link to the collection of identity providers of the cluster.
func (o *Cluster) GetIdentityProviders() (value *IdentityProviderList, ok bool) {
	ok = o != nil && o.identityProviders != nil
	if ok {
		value = o.identityProviders
	}
	return
}

// Ingresses returns the value of the 'ingresses' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// List of ingresses on this cluster.
func (o *Cluster) Ingresses() *IngressList {
	if o == nil {
		return nil
	}
	return o.ingresses
}

// GetIngresses returns the value of the 'ingresses' attribute and
// a flag indicating if the attribute has a value.
//
// List of ingresses on this cluster.
func (o *Cluster) GetIngresses() (value *IngressList, ok bool) {
	ok = o != nil && o.ingresses != nil
	if ok {
		value = o.ingresses
	}
	return
}

// LoadBalancerQuota returns the value of the 'load_balancer_quota' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Load Balancer quota to be assigned to the cluster.
func (o *Cluster) LoadBalancerQuota() int {
	if o != nil && o.loadBalancerQuota != nil {
		return *o.loadBalancerQuota
	}
	return 0
}

// GetLoadBalancerQuota returns the value of the 'load_balancer_quota' attribute and
// a flag indicating if the attribute has a value.
//
// Load Balancer quota to be assigned to the cluster.
func (o *Cluster) GetLoadBalancerQuota() (value int, ok bool) {
	ok = o != nil && o.loadBalancerQuota != nil
	if ok {
		value = *o.loadBalancerQuota
	}
	return
}

// Managed returns the value of the 'managed' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Flag indicating if the cluster is managed (by Red Hat) or
// self-managed by the user.
func (o *Cluster) Managed() bool {
	if o != nil && o.managed != nil {
		return *o.managed
	}
	return false
}

// GetManaged returns the value of the 'managed' attribute and
// a flag indicating if the attribute has a value.
//
// Flag indicating if the cluster is managed (by Red Hat) or
// self-managed by the user.
func (o *Cluster) GetManaged() (value bool, ok bool) {
	ok = o != nil && o.managed != nil
	if ok {
		value = *o.managed
	}
	return
}

// Metrics returns the value of the 'metrics' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Cluster metrics received from telemetry.
//
// When provisioning a cluster this will be ignored.
func (o *Cluster) Metrics() *ClusterMetrics {
	if o == nil {
		return nil
	}
	return o.metrics
}

// GetMetrics returns the value of the 'metrics' attribute and
// a flag indicating if the attribute has a value.
//
// Cluster metrics received from telemetry.
//
// When provisioning a cluster this will be ignored.
func (o *Cluster) GetMetrics() (value *ClusterMetrics, ok bool) {
	ok = o != nil && o.metrics != nil
	if ok {
		value = o.metrics
	}
	return
}

// MultiAZ returns the value of the 'multi_AZ' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Flag indicating if the cluster should be created with nodes in
// different availability zones or all the nodes in a single one
// randomly selected.
func (o *Cluster) MultiAZ() bool {
	if o != nil && o.multiAZ != nil {
		return *o.multiAZ
	}
	return false
}

// GetMultiAZ returns the value of the 'multi_AZ' attribute and
// a flag indicating if the attribute has a value.
//
// Flag indicating if the cluster should be created with nodes in
// different availability zones or all the nodes in a single one
// randomly selected.
func (o *Cluster) GetMultiAZ() (value bool, ok bool) {
	ok = o != nil && o.multiAZ != nil
	if ok {
		value = *o.multiAZ
	}
	return
}

// Name returns the value of the 'name' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Name of the cluster. This name is assigned by the user when the
// cluster is created.
func (o *Cluster) Name() string {
	if o != nil && o.name != nil {
		return *o.name
	}
	return ""
}

// GetName returns the value of the 'name' attribute and
// a flag indicating if the attribute has a value.
//
// Name of the cluster. This name is assigned by the user when the
// cluster is created.
func (o *Cluster) GetName() (value string, ok bool) {
	ok = o != nil && o.name != nil
	if ok {
		value = *o.name
	}
	return
}

// Network returns the value of the 'network' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Network settings of the cluster.
func (o *Cluster) Network() *Network {
	if o == nil {
		return nil
	}
	return o.network
}

// GetNetwork returns the value of the 'network' attribute and
// a flag indicating if the attribute has a value.
//
// Network settings of the cluster.
func (o *Cluster) GetNetwork() (value *Network, ok bool) {
	ok = o != nil && o.network != nil
	if ok {
		value = o.network
	}
	return
}

// Nodes returns the value of the 'nodes' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Information about the nodes of the cluster.
func (o *Cluster) Nodes() *ClusterNodes {
	if o == nil {
		return nil
	}
	return o.nodes
}

// GetNodes returns the value of the 'nodes' attribute and
// a flag indicating if the attribute has a value.
//
// Information about the nodes of the cluster.
func (o *Cluster) GetNodes() (value *ClusterNodes, ok bool) {
	ok = o != nil && o.nodes != nil
	if ok {
		value = o.nodes
	}
	return
}

// OpenshiftVersion returns the value of the 'openshift_version' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Version of _OpenShift_ installed in the cluster, for example `4.0.0-0.2`.
//
// When retrieving a cluster this will always be reported.
//
// When provisioning a cluster this will be ignored, as the version to
// deploy will be determined internally.
func (o *Cluster) OpenshiftVersion() string {
	if o != nil && o.openshiftVersion != nil {
		return *o.openshiftVersion
	}
	return ""
}

// GetOpenshiftVersion returns the value of the 'openshift_version' attribute and
// a flag indicating if the attribute has a value.
//
// Version of _OpenShift_ installed in the cluster, for example `4.0.0-0.2`.
//
// When retrieving a cluster this will always be reported.
//
// When provisioning a cluster this will be ignored, as the version to
// deploy will be determined internally.
func (o *Cluster) GetOpenshiftVersion() (value string, ok bool) {
	ok = o != nil && o.openshiftVersion != nil
	if ok {
		value = *o.openshiftVersion
	}
	return
}

// Product returns the value of the 'product' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Link to the product type of this cluster.
func (o *Cluster) Product() *Product {
	if o == nil {
		return nil
	}
	return o.product
}

// GetProduct returns the value of the 'product' attribute and
// a flag indicating if the attribute has a value.
//
// Link to the product type of this cluster.
func (o *Cluster) GetProduct() (value *Product, ok bool) {
	ok = o != nil && o.product != nil
	if ok {
		value = o.product
	}
	return
}

// Properties returns the value of the 'properties' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// User defined properties for tagging and querying.
func (o *Cluster) Properties() map[string]string {
	if o == nil {
		return nil
	}
	return o.properties
}

// GetProperties returns the value of the 'properties' attribute and
// a flag indicating if the attribute has a value.
//
// User defined properties for tagging and querying.
func (o *Cluster) GetProperties() (value map[string]string, ok bool) {
	ok = o != nil && o.properties != nil
	if ok {
		value = o.properties
	}
	return
}

// ProvisionShard returns the value of the 'provision_shard' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// ProvisionShard contains the properties of the provision shard, including AWS and GCP related configurations
func (o *Cluster) ProvisionShard() *ProvisionShard {
	if o == nil {
		return nil
	}
	return o.provisionShard
}

// GetProvisionShard returns the value of the 'provision_shard' attribute and
// a flag indicating if the attribute has a value.
//
// ProvisionShard contains the properties of the provision shard, including AWS and GCP related configurations
func (o *Cluster) GetProvisionShard() (value *ProvisionShard, ok bool) {
	ok = o != nil && o.provisionShard != nil
	if ok {
		value = o.provisionShard
	}
	return
}

// Region returns the value of the 'region' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Link to the cloud provider region where the cluster is installed.
func (o *Cluster) Region() *CloudRegion {
	if o == nil {
		return nil
	}
	return o.region
}

// GetRegion returns the value of the 'region' attribute and
// a flag indicating if the attribute has a value.
//
// Link to the cloud provider region where the cluster is installed.
func (o *Cluster) GetRegion() (value *CloudRegion, ok bool) {
	ok = o != nil && o.region != nil
	if ok {
		value = o.region
	}
	return
}

// State returns the value of the 'state' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Overall state of the cluster.
func (o *Cluster) State() ClusterState {
	if o != nil && o.state != nil {
		return *o.state
	}
	return ClusterState("")
}

// GetState returns the value of the 'state' attribute and
// a flag indicating if the attribute has a value.
//
// Overall state of the cluster.
func (o *Cluster) GetState() (value ClusterState, ok bool) {
	ok = o != nil && o.state != nil
	if ok {
		value = *o.state
	}
	return
}

// Status returns the value of the 'status' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Status of cluster
func (o *Cluster) Status() *ClusterStatus {
	if o == nil {
		return nil
	}
	return o.status
}

// GetStatus returns the value of the 'status' attribute and
// a flag indicating if the attribute has a value.
//
// Status of cluster
func (o *Cluster) GetStatus() (value *ClusterStatus, ok bool) {
	ok = o != nil && o.status != nil
	if ok {
		value = o.status
	}
	return
}

// StorageQuota returns the value of the 'storage_quota' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Storage quota to be assigned to the cluster.
func (o *Cluster) StorageQuota() *Value {
	if o == nil {
		return nil
	}
	return o.storageQuota
}

// GetStorageQuota returns the value of the 'storage_quota' attribute and
// a flag indicating if the attribute has a value.
//
// Storage quota to be assigned to the cluster.
func (o *Cluster) GetStorageQuota() (value *Value, ok bool) {
	ok = o != nil && o.storageQuota != nil
	if ok {
		value = o.storageQuota
	}
	return
}

// Subscription returns the value of the 'subscription' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Link to the subscription that comes from the account management service when the cluster
// is registered.
func (o *Cluster) Subscription() *Subscription {
	if o == nil {
		return nil
	}
	return o.subscription
}

// GetSubscription returns the value of the 'subscription' attribute and
// a flag indicating if the attribute has a value.
//
// Link to the subscription that comes from the account management service when the cluster
// is registered.
func (o *Cluster) GetSubscription() (value *Subscription, ok bool) {
	ok = o != nil && o.subscription != nil
	if ok {
		value = o.subscription
	}
	return
}

// Version returns the value of the 'version' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Link to the version of _OpenShift_ that will be used to install the cluster.
func (o *Cluster) Version() *Version {
	if o == nil {
		return nil
	}
	return o.version
}

// GetVersion returns the value of the 'version' attribute and
// a flag indicating if the attribute has a value.
//
// Link to the version of _OpenShift_ that will be used to install the cluster.
func (o *Cluster) GetVersion() (value *Version, ok bool) {
	ok = o != nil && o.version != nil
	if ok {
		value = o.version
	}
	return
}

// ClusterListKind is the name of the type used to represent list of objects of
// type 'cluster'.
const ClusterListKind = "ClusterList"

// ClusterListLinkKind is the name of the type used to represent links to list
// of objects of type 'cluster'.
const ClusterListLinkKind = "ClusterListLink"

// ClusterNilKind is the name of the type used to nil lists of objects of
// type 'cluster'.
const ClusterListNilKind = "ClusterListNil"

// ClusterList is a list of values of the 'cluster' type.
type ClusterList struct {
	href  *string
	link  bool
	items []*Cluster
}

// Kind returns the name of the type of the object.
func (l *ClusterList) Kind() string {
	if l == nil {
		return ClusterListNilKind
	}
	if l.link {
		return ClusterListLinkKind
	}
	return ClusterListKind
}

// Link returns true iif this is a link.
func (l *ClusterList) Link() bool {
	return l != nil && l.link
}

// HREF returns the link to the list.
func (l *ClusterList) HREF() string {
	if l != nil && l.href != nil {
		return *l.href
	}
	return ""
}

// GetHREF returns the link of the list and a flag indicating if the
// link has a value.
func (l *ClusterList) GetHREF() (value string, ok bool) {
	ok = l != nil && l.href != nil
	if ok {
		value = *l.href
	}
	return
}

// Len returns the length of the list.
func (l *ClusterList) Len() int {
	if l == nil {
		return 0
	}
	return len(l.items)
}

// Empty returns true if the list is empty.
func (l *ClusterList) Empty() bool {
	return l == nil || len(l.items) == 0
}

// Get returns the item of the list with the given index. If there is no item with
// that index it returns nil.
func (l *ClusterList) Get(i int) *Cluster {
	if l == nil || i < 0 || i >= len(l.items) {
		return nil
	}
	return l.items[i]
}

// Slice returns an slice containing the items of the list. The returned slice is a
// copy of the one used internally, so it can be modified without affecting the
// internal representation.
//
// If you don't need to modify the returned slice consider using the Each or Range
// functions, as they don't need to allocate a new slice.
func (l *ClusterList) Slice() []*Cluster {
	var slice []*Cluster
	if l == nil {
		slice = make([]*Cluster, 0)
	} else {
		slice = make([]*Cluster, len(l.items))
		copy(slice, l.items)
	}
	return slice
}

// Each runs the given function for each item of the list, in order. If the function
// returns false the iteration stops, otherwise it continues till all the elements
// of the list have been processed.
func (l *ClusterList) Each(f func(item *Cluster) bool) {
	if l == nil {
		return
	}
	for _, item := range l.items {
		if !f(item) {
			break
		}
	}
}

// Range runs the given function for each index and item of the list, in order. If
// the function returns false the iteration stops, otherwise it continues till all
// the elements of the list have been processed.
func (l *ClusterList) Range(f func(index int, item *Cluster) bool) {
	if l == nil {
		return
	}
	for index, item := range l.items {
		if !f(index, item) {
			break
		}
	}
}
