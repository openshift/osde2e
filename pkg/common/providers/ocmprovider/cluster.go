package ocmprovider

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os/user"
	"strings"
	"time"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/osde2e/pkg/common/aws"
	"github.com/openshift/osde2e/pkg/common/clusterproperties"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/viper"
)

// IsValidClusterName validates the clustername prior to proceeding with it
// in launching a cluster.
func (o *OCMProvider) IsValidClusterName(clusterName string) (bool, error) {
	// Create a context:
	ctx := context.Background()

	collection := o.conn.ClustersMgmt().V1().Clusters()

	// Retrieve the list of clusters using pages of ten items, till we get a page that has less
	// items than requests, as that marks the end of the collection:
	size := 10
	page := 1
	searchPhrase := fmt.Sprintf("name = '%s'", clusterName)
	for {
		// Retrieve the page:
		response, err := collection.List().
			Search(searchPhrase).
			Size(size).
			Page(page).
			SendContext(ctx)
		if err != nil {
			return false, fmt.Errorf("Can't retrieve page %d: %s\n", page, err)
		}

		if response.Total() != 0 {
			return false, nil
		}

		// Break the loop if the size of the page is less than requested, otherwise go to
		// the next page:
		if response.Size() < size {
			break
		}
		page++
	}

	// Name is valid.
	return true, nil
}

// LaunchCluster setups an new cluster using the OSD API and returns it's ID.
func (o *OCMProvider) LaunchCluster(clusterName string) (string, error) {
	flavourID := getFlavour()
	if flavourID == "" {
		return "", fmt.Errorf("No flavour has been selected")
	}

	// check that enough quota exists for this test if creating cluster
	if enoughQuota, err := o.CheckQuota(flavourID); err != nil {
		log.Printf("Failed to check if enough quota is available: %v", err)
	} else if !enoughQuota {
		return "", fmt.Errorf("currently not enough quota exists to run this test")
	}

	multiAZ := viper.GetBool(config.Cluster.MultiAZ)
	computeMachineType := viper.GetString(ComputeMachineType)
	cloudProvider := viper.GetString(config.CloudProvider.CloudProviderID)

	region, err := o.DetermineRegion(cloudProvider)

	if err != nil {
		return "", fmt.Errorf("error while determining region: %v", err)
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
		Properties(clusterProperties)

	// If AWS credentials are set, this must be a CCS cluster
	awsAccount := viper.GetString(AWSAccount)
	awsAccessKey := viper.GetString(AWSAccessKey)
	awsSecretKey := viper.GetString(AWSSecretKey)

	if awsAccount != "" && awsAccessKey != "" && awsSecretKey != "" {
		newCluster.CCS(v1.NewCCS().Enabled(true)).AWS(
			v1.NewAWS().
				AccountID(awsAccount).
				AccessKeyID(awsAccessKey).
				SecretAccessKey(awsSecretKey))
	}

	expiryInMinutes := viper.GetDuration(config.Cluster.ExpiryInMinutes)
	if expiryInMinutes > 0 {
		// Calculate an expiration date for the cluster so that it will be automatically deleted if
		// we happen to forget to do it:
		expiration := time.Now().Add(expiryInMinutes * time.Minute).UTC() // UTC() to workaround SDA-1567.
		newCluster = newCluster.ExpirationTimestamp(expiration)
	}

	numComputeNodes := viper.GetInt(config.Cluster.NumWorkerNodes)
	if numComputeNodes > 0 {
		nodeBuilder = nodeBuilder.Compute(numComputeNodes)
	}

	// Configure the cluster to be Multi-AZ if configured
	// We must manually configure the number of compute nodes
	// Currently set to 9 nodes. Whatever it is, must be divisible by 3.
	if multiAZ {
		nodeBuilder = nodeBuilder.Compute(9)
		if numComputeNodes > 0 && math.Mod(float64(numComputeNodes), float64(3)) == 0 {
			nodeBuilder = nodeBuilder.Compute(numComputeNodes)
		}
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

// DetermineRegion will return the region provided by configs. This mainly wraps the random functionality for use
// by the MOA provider.
func (o *OCMProvider) DetermineRegion(cloudProvider string) (string, error) {
	region := viper.GetString(config.CloudProvider.Region)

	// If a region is set to "random", it will poll OCM for all the regions available
	// It then will pull a random entry from the list of regions and set the ID to that
	if region == "random" {
		regionsClient := o.conn.ClustersMgmt().V1().CloudProviders().CloudProvider(cloudProvider).Regions().List()

		regions, err := regionsClient.Send()
		if err != nil {
			return "", err
		}

		for range regions.Items().Slice() {
			regionObj := regions.Items().Slice()[rand.Intn(regions.Total())]
			region = regionObj.ID()

			if regionObj.Enabled() {
				break
			}
		}

		log.Printf("Random region requested, selected %s region.", region)

		// Update the Config with the selected random region
		viper.Set(config.CloudProvider.Region, region)
	}
	return region, nil
}

// GenerateProperties will generate a set of properties to assign to a cluster.
func (o *OCMProvider) GenerateProperties() (map[string]string, error) {
	var username string

	// If JobID is not equal to -1, then we're running on prow.
	if viper.GetInt(config.JobID) != -1 {
		username = "prow"
	} else if viper.GetString(UserOverride) != "" {
		username = viper.GetString(UserOverride)
	} else {

		user, err := user.Current()

		if err != nil {
			return nil, fmt.Errorf("unable to get current user: %v", err)
		}

		username = user.Username
	}

	installedversion := viper.GetString(config.Cluster.Version)

	provisionshardID := viper.GetString(config.Cluster.ProvisionShardID)

	properties := map[string]string{
		clusterproperties.MadeByOSDe2e:     "true",
		clusterproperties.OwnedBy:          username,
		clusterproperties.InstalledVersion: installedversion,
		clusterproperties.UpgradeVersion:   "--",
		clusterproperties.Status:           clusterproperties.StatusProvisioning,
	}

	if provisionshardID != "" {
		properties[clusterproperties.ProvisionShardID] = provisionshardID
	}

	additionalLabels := viper.GetString(AdditionalLabels)
	if len(additionalLabels) > 0 {
		for _, label := range strings.Split(additionalLabels, ",") {
			properties[label] = "true"
		}
	}

	jobName := viper.GetString(config.JobName)
	jobID := viper.GetString(config.JobID)

	if jobName != "" {
		properties[clusterproperties.JobName] = jobName
	}

	if jobID != "" {
		properties[clusterproperties.JobID] = jobID
	}

	return properties, nil
}

// DeleteCluster requests the deletion of clusterID.
func (o *OCMProvider) DeleteCluster(clusterID string) error {
	var resp *v1.ClusterDeleteResponse
	var cluster *spi.Cluster
	var err error
	var ok bool

	if cluster, ok = o.clusterCache[clusterID]; !ok {
		cluster, err = o.GetCluster(clusterID)
		if err != nil {
			return fmt.Errorf("error retrieving cluster for deletion: %v", err)
		}
	}

	err = o.AddProperty(cluster, clusterproperties.Status, clusterproperties.StatusUninstalling)
	if err != nil {
		return fmt.Errorf("error adding uninstalling status to cluster: %v", err)
	}

	err = retryer().Do(func() error {
		var err error
		resp, err = o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).
			Delete().
			Send()

		if err != nil {
			return fmt.Errorf("couldn't delete cluster '%s': %v", clusterID, err)
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

		o.updateClusterCache(clusterID, resp.Body())

		return nil
	})

	if err != nil {
		return err
	}

	if resp.Error() != nil {
		return resp.Error()
	}

	log.Printf("Cluster successfully scaled to %d nodes", numComputeNodes)

	return nil
}

// ListClusters returns a list of clusters filtered on key/value pairs
func (o *OCMProvider) ListClusters(query string) ([]*spi.Cluster, error) {
	var clusters []*spi.Cluster

	totalItems := math.MaxInt64
	emptyPage := false
	page := 1

	for len(clusters) < totalItems || emptyPage {
		clusterListRequest := o.conn.ClustersMgmt().V1().Clusters().List()

		response, err := clusterListRequest.Search(query).Page(page).Send()

		if err != nil {
			return nil, err
		}

		if response.Size() == 0 {
			emptyPage = true
		}

		if totalItems == math.MaxInt64 {
			totalItems = response.Total()
		}

		for _, cluster := range response.Items().Slice() {
			spiCluster, err := o.ocmToSPICluster(cluster)
			if err != nil {
				return nil, err
			}
			clusters = append(clusters, spiCluster)
		}

		page++
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
	o.clusterCache[clusterID] = cluster

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
	if creds, ok := o.credentialCache[clusterID]; ok {
		return []byte(creds), nil
	}

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

	o.credentialCache[clusterID] = resp.Body().Kubeconfig()

	return []byte(resp.Body().Kubeconfig()), nil
}

// GetMetrics gathers metrics from OCM on a cluster
func (o *OCMProvider) GetMetrics(clusterID string) (*v1.ClusterMetrics, error) {
	var err error

	clusterClient := o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID)

	cluster, err := clusterClient.Get().Send()
	if err != nil {
		return nil, err
	}

	return cluster.Body().Metrics(), nil
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

				o.updateClusterCache(clusterID, aoar.Body().Cluster())

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

		o.updateClusterCache(clusterID, resp.Body())

		return nil
	})

	if err != nil {
		return err
	}

	if resp.Error() != nil {
		return resp.Error()
	}

	log.Println("Successfully extended cluster expiry time to", extendexpirytime.UTC().Format("Monday, 02 Jan 2006 15:04:05 MST"))

	return nil
}

// AddProperty adds a new property to the properties field of an existing cluster
func (o *OCMProvider) AddProperty(cluster *spi.Cluster, tag string, value string) error {
	var resp *v1.ClusterUpdateResponse

	clusterproperties := cluster.Properties()

	// Apparently, if cluster properties are empty in OCM, the clusterproperties are nil, In this case, we'll just make our own
	// properties map.
	if clusterproperties == nil {
		clusterproperties = map[string]string{}
	}

	clusterproperties[tag] = value

	modifiedCluster, err := v1.NewCluster().Properties(clusterproperties).Build()

	if err != nil {
		return fmt.Errorf("error while building updated modified cluster object with new property: %v", err)
	}

	propertyFilename := fmt.Sprintf("%s.osde2e-cluster-property-update.metrics.prom", cluster.ID())
	data := fmt.Sprintf("# TYPE cicd_cluster_properties gauge\ncicd_cluster_properties{cluster_id=\"%s\",environment=\"%s\",job_id=\"%s\",property=\"%s\",region=\"%s\",value=\"%s\",version=\"%s\"} 0\n", cluster.ID(), o.Environment(), viper.GetString(config.JobID), tag, cluster.Region(), value, cluster.Version())
	log.Println(data)
	aws.WriteToS3(aws.CreateS3URL(viper.GetString(config.Tests.MetricsBucket), "incoming", propertyFilename), []byte(data))

	err = retryer().Do(func() error {
		var err error
		resp, err = o.conn.ClustersMgmt().V1().Clusters().Cluster(cluster.ID()).Update().
			Body(modifiedCluster).
			Send()

		if err != nil {
			err = fmt.Errorf("couldn't update cluster '%s': %v", cluster.ID(), err)
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

	// We need to update the cache post-update
	o.updateClusterCache(cluster.ID(), resp.Body())

	log.Printf("Successfully added property[%s] - %s \n", tag, resp.Body().Properties()[tag])

	return nil
}

// Upgrade initiates a cluster upgrade to the given version
func (o *OCMProvider) Upgrade(clusterID string, version string, pdbTimeoutMinutes int, t time.Time) error {

	nodeDrain := v1.NewValue().Value(float64(pdbTimeoutMinutes)).Unit("minutes")
	policy, err := v1.NewUpgradePolicy().Version(version).NextRun(t).ScheduleType("manual").NodeDrainGracePeriod(nodeDrain).Build()
	if err != nil {
		return err
	}

	addResp, err := o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).UpgradePolicies().Add().Body(policy).SendContext(context.TODO())
	if err != nil {
		return err
	}
	if addResp.Status() != http.StatusCreated {
		log.Printf("Unable to schedule upgrade with provider (status %d, response %v)", addResp.Status(), addResp.Error())
		return err
	}

	log.Printf("upgrade to version %s scheduled with provider for time %s", addResp.Body().Version(), addResp.Body().NextRun().Format(time.RFC3339))
	return nil
}

// This assumes cluster is a resp.Body() response from an OCM update
func (o *OCMProvider) updateClusterCache(id string, cluster *v1.Cluster) error {
	c, err := o.ocmToSPICluster(cluster)
	if err != nil {
		return err
	}
	o.clusterCache[cluster.ID()] = c
	return nil
}
