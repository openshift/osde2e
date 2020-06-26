package ocmprovider

import (
	"fmt"
	"log"
	"math/rand"
	"os/user"
	"strings"
	"time"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/viper"
)

const (
	// DefaultFlavour is used when no specialized configuration exists.
	DefaultFlavour = "osd-4"

	// MadeByOSDe2e property to attach to clusters.
	MadeByOSDe2e = "MadeByOSDe2e"

	// OwnedBy property which will tell who made the cluster.
	OwnedBy = "OwnedBy"
)

// LaunchCluster setups an new cluster using the OSD API and returns it's ID.
func (o *OCMProvider) LaunchCluster() (string, error) {
	clusterName := viper.GetString(config.Cluster.Name)
	log.Printf("Creating cluster '%s'...", clusterName)

	// choose flavour based on config
	flavourID := DefaultFlavour

	// Calculate an expiration date for the cluster so that it will be automatically deleted if
	// we happen to forget to do it:
	expiration := time.Now().Add(time.Duration(viper.GetInt64(config.Cluster.ExpiryInMinutes)) * time.Minute).UTC() // UTC() to workaround SDA-1567.

	multiAZ := viper.GetBool(config.Cluster.MultiAZ)
	computeMachineType := viper.GetString(ComputeMachineType)
	region := viper.GetString(config.CloudProvider.Region)
	cloudProvider := viper.GetString(config.CloudProvider.CloudProviderID)

	// If a region is set to "random", it will poll OCM for all the regions available
	// It then will pull a random entry from the list of regions and set the ID to that
	if region == "random" {
		regionsClient := o.conn.ClustersMgmt().V1().CloudProviders().CloudProvider(cloudProvider).Regions().List()
		regions, err := regionsClient.Send()
		if err != nil {
			return "", err
		}

		region = regions.Items().Slice()[rand.Intn(regions.Total())].ID()
	}

	nodeBuilder := &v1.ClusterNodesBuilder{}

	clusterProperties, err := o.GenerateProperties()

	if err != nil {
		return "", fmt.Errorf("error generating cluster properties: %v", err)
	}

	newCluster := v1.NewCluster().
		Name(clusterName).
		Flavour(v1.NewFlavour().
			ID(flavourID)).
		Region(v1.NewCloudRegion().
			ID(region)).
		MultiAZ(multiAZ).
		Version(v1.NewVersion().
			ID(viper.GetString(config.Cluster.Version))).
		CloudProvider(v1.NewCloudProvider().
			ID(cloudProvider)).
		ExpirationTimestamp(expiration).
		Properties(clusterProperties)

	// Configure the cluster to be Multi-AZ if configured
	// We must manually configure the number of compute nodes
	// Currently set to 9 nodes. Whatever it is, must be divisible by 3.
	if multiAZ {
		nodeBuilder = nodeBuilder.Compute(9)
		newCluster = newCluster.MultiAZ(viper.GetBool(config.Cluster.MultiAZ))
	}

	if computeMachineType != "" {
		machineType := &v1.MachineTypeBuilder{}
		nodeBuilder = nodeBuilder.ComputeMachineType(machineType.ID(computeMachineType))
	}

	newCluster = newCluster.Nodes(nodeBuilder)

	IDsAtCreationString := viper.GetString(config.Addons.IDsAtCreation)
	if len(IDsAtCreationString) > 0 {
		addons := []*v1.AddOnInstallationBuilder{}
		IDsAtCreation := strings.Split(IDsAtCreationString, ",")
		for _, id := range IDsAtCreation {
			addons = append(addons, v1.NewAddOnInstallation().Addon(v1.NewAddOn().ID(id)))
		}

		newCluster = newCluster.Addons(v1.NewAddOnInstallationList().Items(addons...))
	}

	cluster, err := newCluster.Build()
	if err != nil {
		return "", fmt.Errorf("couldn't build cluster description: %v", err)
	}

	var resp *v1.ClustersAddResponse

	err = retryer().Do(func() error {
		var err error
		resp, err = o.conn.ClustersMgmt().V1().Clusters().Add().
			Body(cluster).
			Send()

		if resp != nil && resp.Error() != nil {
			return errResp(resp.Error())
		}

		return err
	})

	if err != nil {
		return "", fmt.Errorf("couldn't create cluster: %v", err)
	}
	return resp.Body().ID(), nil
}

func (o *OCMProvider) GenerateProperties() (map[string]string, error) {
	var username string

	// If JobID is not equal to -1, then we're running on prow.
	if viper.GetInt(config.JobID) != -1 {
		username = "prow"
	} else {

		user, err := user.Current()

		if err != nil {
			return nil, fmt.Errorf("unable to get current user: %v", err)
		}

		username = user.Username
	}

	return map[string]string{
		MadeByOSDe2e: "true",
		OwnedBy:      username,
	}, nil
}

// DeleteCluster requests the deletion of clusterID.
func (o *OCMProvider) DeleteCluster(clusterID string) error {
	var resp *v1.ClusterDeleteResponse

	err := retryer().Do(func() error {
		var err error
		resp, err = o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).
			Delete().
			Send()

		if err != nil {
			log.Printf("couldn't delete cluster: %v", err)
			return err
		}

		if resp != nil && resp.Error() != nil {
			err = errResp(resp.Error())
			log.Printf("%v", err)
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("couldn't delete cluster '%s': %v", clusterID, err)
	}
	return nil
}

// ScaleCluster will grow or shink the cluster to the desired number of compute nodes.
func (o *OCMProvider) ScaleCluster(clusterID string, numComputeNodes int) error {
	var resp *v1.ClusterUpdateResponse

	// Get the current state of the cluster
	ocmCluster, err := o.getOCMCluster(clusterID)

	if err != nil {
		return err
	}

	if numComputeNodes == ocmCluster.Nodes().Compute() {
		log.Printf("cluster already at desired size (%d)", numComputeNodes)
		return nil
	}

	scaledCluster, err := v1.NewCluster().
		Nodes(v1.NewClusterNodes().
			Compute(numComputeNodes)).
		Build()

	if err != nil {
		return fmt.Errorf("error while building scaled cluster object: %v", err)
	}

	err = retryer().Do(func() error {
		var err error
		resp, err = o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).Update().
			Body(scaledCluster).
			Send()

		if err != nil {
			err = fmt.Errorf("couldn't update cluster '%s': %v", clusterID, err)
			log.Printf("%v", err)
			return err
		}

		if resp != nil && resp.Error() != nil {
			log.Printf("error while trying to update cluster: %v", err)
			return errResp(resp.Error())
		}

		return nil
	})

	if err != nil {
		return err
	}

	if resp.Error() != nil {
		return resp.Error()
	}

	finalCluster, err := o.GetCluster(clusterID)

	if err != nil {
		log.Printf("error attempting to retrieve cluster for verification: %v", err)
	}

	if finalCluster.NumComputeNodes() != numComputeNodes {
		return fmt.Errorf("expected number of compute nodes (%d) not reflected in OCM (found %d)", numComputeNodes, finalCluster.NumComputeNodes())

	}
	log.Printf("Cluster successfully scaled to %d nodes", numComputeNodes)

	return nil
}

// ListClusters returns a list of clusters filtered on key/value pairs
func (o *OCMProvider) ListClusters(query string) ([]*spi.Cluster, error) {
	var clusters []*spi.Cluster
	clusterListRequest := o.conn.ClustersMgmt().V1().Clusters().List()

	response, err := clusterListRequest.Search(query).Send()

	if err != nil {
		return nil, err
	}

	for _, cluster := range response.Items().Slice() {
		spiCluster, err := o.ocmToSPICluster(cluster)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, spiCluster)
	}

	return clusters, nil
}

// GetCluster returns a cluster from OCM.
func (o *OCMProvider) GetCluster(clusterID string) (*spi.Cluster, error) {
	ocmCluster, err := o.getOCMCluster(clusterID)
	if err != nil {
		return nil, err
	}

	cluster, err := o.ocmToSPICluster(ocmCluster)
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

func (o *OCMProvider) getOCMCluster(clusterID string) (*v1.Cluster, error) {
	var resp *v1.ClusterGetResponse

	err := retryer().Do(func() error {
		var err error
		resp, err = o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).
			Get().
			Send()

		if err != nil {
			err = fmt.Errorf("couldn't retrieve cluster '%s': %v", clusterID, err)
			log.Printf("%v", err)
			return err
		}

		if resp != nil && resp.Error() != nil {
			log.Printf("error while trying to retrieve cluster: %v", err)
			return errResp(resp.Error())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if resp.Error() != nil {
		return nil, resp.Error()
	}

	return resp.Body(), nil
}

// ClusterKubeconfig returns the kubeconfig for the given cluster ID.
func (o *OCMProvider) ClusterKubeconfig(clusterID string) ([]byte, error) {
	var resp *v1.CredentialsGetResponse

	err := retryer().Do(func() error {
		var err error
		resp, err = o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).
			Credentials().
			Get().
			Send()

		if err != nil {
			log.Printf("couldn't get credentials: %v", err)
			return err
		}

		if resp != nil && resp.Error() != nil {
			err = errResp(resp.Error())
			log.Printf("%v", err)
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve credentials for cluster '%s': %v", clusterID, err)
	}
	return []byte(resp.Body().Kubeconfig()), nil
}

// GetMetrics gathers metrics from OCM on a cluster
func (o *OCMProvider) GetMetrics(clusterID string) (*v1.ClusterMetrics, error) {
	var err error
	var alertsMetricQuery *v1.AlertsMetricQueryGetResponse
	var operatorsMetricQuery *v1.ClusterOperatorsMetricQueryGetResponse
	var nodesMetricQuery *v1.NodesMetricQueryGetResponse
	var socketTotalClient *v1.SocketTotalByNodeRolesOSMetricQueryGetResponse
	var CPUTotalClient *v1.CPUTotalByNodeRolesOSMetricQueryGetResponse

	clusterMetricsBuilder := v1.NewClusterMetrics()

	clusterClient := o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID)

	metricsClient := clusterClient.MetricQueries()

	err = retryer().Do(func() error {
		alertsMetricQuery, err = metricsClient.Alerts().Get().Send()
		return err
	})
	if err != nil {
		return nil, err
	}

	criticalAlerts := 0
	for _, alert := range alertsMetricQuery.Body().Alerts() {
		if alert.Severity() == v1.AlertSeverityCritical {
			criticalAlerts++
		}
	}
	clusterMetricsBuilder.CriticalAlertsFiring(criticalAlerts)

	err = retryer().Do(func() error {
		operatorsMetricQuery, err = metricsClient.ClusterOperators().Get().Send()
		return err
	})
	if err != nil {
		return nil, err
	}

	failingOperators := 0
	for _, operator := range operatorsMetricQuery.Body().Operators() {
		if operator.Condition() == v1.ClusterOperatorStateFailing {
			failingOperators++
		}
	}
	clusterMetricsBuilder.OperatorsConditionFailing(failingOperators)

	err = retryer().Do(func() error {
		nodesMetricQuery, err = metricsClient.Nodes().Get().Send()
		return err
	})
	if err != nil {
		return nil, err
	}

	infraNodes := 0
	computeNodes := 0
	masterNodes := 0

	for _, node := range nodesMetricQuery.Body().Nodes() {
		node.Amount()
		switch node.Type() {
		case v1.NodeTypeCompute:
			computeNodes = node.Amount()
		case v1.NodeTypeInfra:
			infraNodes = node.Amount()
		case v1.NodeTypeMaster:
			masterNodes = node.Amount()
		}
	}

	clusterMetricsBuilder.Nodes(v1.NewClusterNodes().
		Compute(computeNodes).
		Infra(infraNodes).
		Master(masterNodes).
		Total(computeNodes + infraNodes + masterNodes))

	err = retryer().Do(func() error {
		socketTotalClient, err = metricsClient.SocketTotalByNodeRolesOS().Get().Send()
		return err
	})
	if err != nil {
		return nil, err
	}

	for _, sockets := range socketTotalClient.Body().SocketTotals() {
		for _, role := range sockets.NodeRoles() {
			if role == fmt.Sprintf("%v", v1.NodeTypeCompute) {
				metricBuilder := v1.ClusterMetricBuilder{}
				value := v1.ValueBuilder{}
				value.Value(sockets.SocketTotal())
				value.Unit("sockets")
				metricBuilder.Total(&value)
				clusterMetricsBuilder.ComputeNodesSockets(&metricBuilder)
			}
		}
	}

	err = retryer().Do(func() error {
		CPUTotalClient, err = metricsClient.CPUTotalByNodeRolesOS().Get().Send()
		return err
	})
	if err != nil {
		return nil, err
	}

	for _, cpu := range CPUTotalClient.Body().CPUTotals() {
		for _, role := range cpu.NodeRoles() {
			if role == fmt.Sprintf("%v", v1.NodeTypeCompute) {
				metricBuilder := v1.ClusterMetricBuilder{}
				value := v1.ValueBuilder{}
				value.Value(cpu.CPUTotal())
				value.Unit("cpu")
				metricBuilder.Total(&value)
				clusterMetricsBuilder.ComputeNodesCPU(&metricBuilder)
			}
		}
	}

	clusterMetrics, err := clusterMetricsBuilder.Build()
	if err != nil {
		return nil, err
	}

	return clusterMetrics, nil
}

// InstallAddons loops through the addons list in the config
// and performs the CRUD operation to trigger addon installation
func (o *OCMProvider) InstallAddons(clusterID string, addonIDs []string) (num int, err error) {
	num = 0
	addonsClient := o.conn.ClustersMgmt().V1().Addons()
	clusterClient := o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID)
	for _, addonID := range addonIDs {
		var addonResp *v1.AddOnGetResponse

		err = retryer().Do(func() error {
			var err error
			addonResp, err = addonsClient.Addon(addonID).Get().Send()

			if err != nil {
				return err
			}

			if addonResp != nil && addonResp.Error() != nil {
				return errResp(addonResp.Error())
			}

			return nil
		})

		if err != nil {
			return 0, err
		}
		addon := addonResp.Body()

		alreadyInstalled := false
		cluster, err := o.GetCluster(clusterID)

		if err != nil {
			return 0, fmt.Errorf("error getting current cluster state when trying to install addon %s", addonID)
		}

		for _, addon := range cluster.Addons() {
			if addon == addonID {
				alreadyInstalled = true
				break
			}
		}

		if alreadyInstalled {
			log.Printf("Addon %s is already installed. Skipping.", addonID)
			continue
		}

		if addon.Enabled() {
			addonInstallation, err := v1.NewAddOnInstallation().Addon(v1.NewAddOn().Copy(addon)).Build()
			if err != nil {
				return 0, err
			}

			var aoar *v1.AddOnInstallationsAddResponse

			err = retryer().Do(func() error {
				var err error
				aoar, err = clusterClient.Addons().Add().Body(addonInstallation).Send()
				if err != nil {
					log.Printf("couldn't install addons: %v", err)
					return err
				}

				if aoar.Error() != nil {
					err = fmt.Errorf("error (%v) sending request: %v", aoar.Status(), aoar.Error())
					log.Printf("%v", err)
					return err
				}

				return nil
			})

			if err != nil {
				return 0, err
			}

			log.Printf("Installed Addon: %s", addonID)

			num++
		}
	}

	return num, nil
}

func (o *OCMProvider) ocmToSPICluster(ocmCluster *v1.Cluster) (*spi.Cluster, error) {
	var err error
	var resp *v1.ClusterGetResponse

	cluster := spi.NewClusterBuilder().
		Name(ocmCluster.Name()).
		Region(ocmCluster.Region().ID()).
		Flavour(ocmCluster.Flavour().ID())

	if id, ok := ocmCluster.GetID(); ok {
		cluster.ID(id)
	}

	if version, ok := ocmCluster.GetVersion(); ok {
		cluster.Version(version.ID())
	}

	if cloudProvider, ok := ocmCluster.GetCloudProvider(); ok {
		cluster.CloudProvider(cloudProvider.ID())
	}

	if state, ok := ocmCluster.GetState(); ok {
		cluster.State(ocmStateToInternalState(state))
	}

	if properties, ok := ocmCluster.GetProperties(); ok {
		cluster.Properties(properties)
	}

	var addonsResp *v1.AddOnInstallationsListResponse
	err = retryer().Do(func() error {
		var err error
		addonsResp, err = o.conn.ClustersMgmt().V1().Clusters().Cluster(ocmCluster.ID()).Addons().
			List().
			Send()

		if err != nil {
			err = fmt.Errorf("couldn't retrieve addons for cluster '%s': %v", ocmCluster.ID(), err)
			log.Printf("%v", err)
			return err
		}

		if addonsResp != nil && addonsResp.Error() != nil {
			log.Printf("error while trying to retrieve addons list for cluster: %v", err)
			return errResp(resp.Error())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if addonsResp.Error() != nil {
		return nil, addonsResp.Error()
	}

	if addons, ok := addonsResp.GetItems(); ok {
		addons.Each(func(addon *v1.AddOnInstallation) bool {
			cluster.AddAddon(addon.ID())
			return true
		})
	}

	cluster.ExpirationTimestamp(ocmCluster.ExpirationTimestamp())
	cluster.NumComputeNodes(ocmCluster.Nodes().Compute())

	return cluster.Build(), nil
}

func ocmStateToInternalState(state v1.ClusterState) spi.ClusterState {
	switch state {
	case v1.ClusterStateError:
		return spi.ClusterStateError
	case v1.ClusterStateInstalling:
		return spi.ClusterStateInstalling
	case v1.ClusterStatePending:
		return spi.ClusterStatePending
	case v1.ClusterStateReady:
		return spi.ClusterStateReady
	case v1.ClusterStateUninstalling:
		return spi.ClusterStateUninstalling
	default:
		return spi.ClusterStateUnknown
	}
}

// ExtendExpiry extends the expiration time of an existing cluster
func (o *OCMProvider) ExtendExpiry(clusterID string, hours uint64, minutes uint64, seconds uint64) error {
	var resp *v1.ClusterUpdateResponse

	// Get the current state of the cluster
	ocmCluster, err := o.getOCMCluster(clusterID)

	if err != nil {
		return err
	}

	cluster, err := o.ocmToSPICluster(ocmCluster)
	if err != nil {
		return err
	}

	extendexpirytime := cluster.ExpirationTimestamp()

	if hours != 0 {
		extendexpirytime = extendexpirytime.Add(time.Duration(hours) * time.Hour).UTC()
	}
	if minutes != 0 {
		extendexpirytime = extendexpirytime.Add(time.Duration(minutes) * time.Minute).UTC()
	}
	if seconds != 0 {
		extendexpirytime = extendexpirytime.Add(time.Duration(seconds) * time.Second).UTC()
	}

	extendexpiryCluster, err := v1.NewCluster().ExpirationTimestamp(extendexpirytime).Build()

	if err != nil {
		return fmt.Errorf("error while building updated expiration time cluster object: %v", err)
	}

	err = retryer().Do(func() error {
		var err error
		resp, err = o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).Update().
			Body(extendexpiryCluster).
			Send()

		if err != nil {
			err = fmt.Errorf("couldn't update cluster '%s': %v", clusterID, err)
			log.Printf("%v", err)
			return err
		}

		if resp != nil && resp.Error() != nil {
			log.Printf("error while trying to update cluster: %v", err)
			return errResp(resp.Error())
		}

		return nil
	})

	if err != nil {
		return err
	}

	if resp.Error() != nil {
		return resp.Error()
	}

	finalCluster, err := o.GetCluster(clusterID)

	if err != nil {
		log.Printf("error attempting to retrieve cluster for verification: %v", err)
	}

	if finalCluster.ExpirationTimestamp() != extendexpirytime {
		return fmt.Errorf("expected expiration time %s not reflected in OCM (found %s)", extendexpirytime.UTC().Format("2002-01-02 14:03:02 Monday"), finalCluster.ExpirationTimestamp().UTC().Format("2002-01-02 14:03:02 Monday"))

	}
	log.Println("Successfully extended cluster expiry time to ", extendexpirytime.UTC().Format("2002-01-02 14:03:02 Monday"))

	return nil
}
