package spi

import (
	"time"

	"github.com/openshift/osde2e/pkg/common/metadata"
)

// ClusterState is the state of the cluster.
type ClusterState string

const (
	// ClusterStateError is the cluster error state.
	ClusterStateError ClusterState = "error"
	// ClusterStateInstalling the cluster is being isntalled.
	ClusterStateInstalling ClusterState = "installing"
	// ClusterStatePending the cluster is awaiting installation.
	ClusterStatePending ClusterState = "pending"
	// ClusterStatePendingAccount the cluster is awaiting an account
	ClusterStatePendingAccount ClusterState = "pending_account"
	// ClusterStateReady the cluster is ready to use.
	ClusterStateReady ClusterState = "ready"
	// ClusterStateUninstalling the cluster is uninstalling.
	ClusterStateUninstalling ClusterState = "uninstalling"
	// ClusterStateUnknown the cluster state is unknown.
	ClusterStateUnknown ClusterState = "unknown"
	// ClusterStatePoweringDown
	ClusterStatePoweringDown ClusterState = "powering_down"
	// ClusterStateHibernating
	ClusterStateHibernating ClusterState = "hibernating"
	// ClusterStateResuming
	ClusterStateResuming ClusterState = "resuming"
)

// Cluster is the intermediary cluster object between a provisioner and osde2e.
type Cluster struct {
	id                  string
	name                string
	version             string
	cloudProvider       string
	product             string
	region              string
	creationTimestamp   time.Time
	expirationTimestamp time.Time
	state               ClusterState
	flavour             string
	addons              []string
	numComputeNodes     int
	// We want to be provider-agnostic,  butevery cluster we interact with
	// should provide us these metrics. The clustersmgmt metrics types handle
	// all aspects of this extremely well, so let's not reinvent the wheel.

	// Custom tags for a cluster are included in the properties field
	properties map[string]string
}

// ID returns the cluster ID.
func (c *Cluster) ID() string {
	return c.id
}

// Name returns the cluster name.
func (c *Cluster) Name() string {
	return c.name
}

// Version returns the cluster version.
func (c *Cluster) Version() string {
	return c.version
}

// CloudProvider returns the cloud provider.
func (c *Cluster) CloudProvider() string {
	return c.cloudProvider
}

// Product returns the product.
func (c *Cluster) Product() string {
	return c.product
}

// Region returns the cloud provider region.
func (c *Cluster) Region() string {
	return c.region
}

// ExpirationTimestamp returns the expiration timestamp.
func (c *Cluster) ExpirationTimestamp() time.Time {
	return c.expirationTimestamp
}

// CreationTimestamp returns the creation timestamp.
func (c *Cluster) CreationTimestamp() time.Time {
	return c.creationTimestamp
}

// State returns the cluster state.
func (c *Cluster) State() ClusterState {
	return c.state
}

// Flavour returns the cluster flavour.
func (c *Cluster) Flavour() string {
	return c.flavour
}

// Addons returns the list of cluster addons.
func (c *Cluster) Addons() []string {
	return c.addons
}

// NumComputeNodes returns the number of compute nodes.
func (c *Cluster) NumComputeNodes() int {
	return c.numComputeNodes
}

// Properties returns properties related to the given cluster.
func (c *Cluster) Properties() map[string]string {
	return c.properties
}

// ClusterBuilder is a struct that can create cluster objects.
type ClusterBuilder struct {
	id                  string
	name                string
	version             string
	cloudProvider       string
	product             string
	region              string
	creationTimestamp   time.Time
	expirationTimestamp time.Time
	state               ClusterState
	flavour             string
	addons              []string
	numComputeNodes     int
	properties          map[string]string
}

// NewClusterBuilder creates a new cluster builder that can create a new cluster.
func NewClusterBuilder() *ClusterBuilder {
	return &ClusterBuilder{
		state:  ClusterStateUnknown,
		addons: []string{},
	}
}

// ID sets the ID for a cluster builder.
func (cb *ClusterBuilder) ID(id string) *ClusterBuilder {
	cb.id = id
	return cb
}

// Name sets the name for a cluster builder.
func (cb *ClusterBuilder) Name(name string) *ClusterBuilder {
	cb.name = name
	return cb
}

// Version sets the version for a cluster builder.
func (cb *ClusterBuilder) Version(version string) *ClusterBuilder {
	cb.version = version
	return cb
}

// CloudProvider sets the cloud provider for a cluster builder.
func (cb *ClusterBuilder) CloudProvider(cloudProvider string) *ClusterBuilder {
	cb.cloudProvider = cloudProvider
	return cb
}

// Product sets the cloud provider for a cluster builder.
func (cb *ClusterBuilder) Product(product string) *ClusterBuilder {
	cb.product = product
	return cb
}

// Region sets the region for a cluster builder.
func (cb *ClusterBuilder) Region(region string) *ClusterBuilder {
	cb.region = region
	return cb
}

// ExpirationTimestamp sets the expiration timestamp for a cluster builder.
func (cb *ClusterBuilder) ExpirationTimestamp(expirationTimestamp time.Time) *ClusterBuilder {
	cb.expirationTimestamp = expirationTimestamp
	return cb
}

// CreationTimestamp sets the creation timestamp for a cluster builder.
func (cb *ClusterBuilder) CreationTimestamp(creationTimestamp time.Time) *ClusterBuilder {
	cb.creationTimestamp = creationTimestamp
	return cb
}

// State sets the state for a cluster builder.
func (cb *ClusterBuilder) State(state ClusterState) *ClusterBuilder {
	cb.state = state
	metadata.Instance.SetStatus(string(state))
	return cb
}

// Flavour sets the flavour for a cluster builder.
func (cb *ClusterBuilder) Flavour(flavour string) *ClusterBuilder {
	cb.flavour = flavour
	return cb
}

// Addons sets the list of addons for a cluster builder.
func (cb *ClusterBuilder) Addons(addons []string) *ClusterBuilder {
	cb.addons = addons
	return cb
}

// AddAddon appends the addon to the list of addons.
func (cb *ClusterBuilder) AddAddon(addon string) *ClusterBuilder {
	cb.addons = append(cb.addons, addon)
	return cb
}

// NumComputeNodes sets the number of compute nodes
func (cb *ClusterBuilder) NumComputeNodes(numComputeNodes int) *ClusterBuilder {
	cb.numComputeNodes = numComputeNodes
	return cb
}

// Properties sets the list of properties for a cluster builder.
func (cb *ClusterBuilder) Properties(properties map[string]string) *ClusterBuilder {
	cb.properties = properties
	return cb
}

// Build will create the cluster from the cluster build.
func (cb *ClusterBuilder) Build() *Cluster {
	return &Cluster{
		id:                  cb.id,
		name:                cb.name,
		version:             cb.version,
		cloudProvider:       cb.cloudProvider,
		product:             cb.product,
		region:              cb.region,
		expirationTimestamp: cb.expirationTimestamp,
		creationTimestamp:   cb.creationTimestamp,
		state:               cb.state,
		flavour:             cb.flavour,
		addons:              cb.addons,
		numComputeNodes:     cb.numComputeNodes,
		properties:          cb.properties,
	}
}
