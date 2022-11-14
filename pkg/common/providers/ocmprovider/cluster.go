package ocmprovider

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/osde2e/pkg/common/aws"
	"github.com/openshift/osde2e/pkg/common/clusterproperties"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"k8s.io/apimachinery/pkg/util/wait"
)

// IsValidClusterName validates the clustername prior to proceeding with it
// in launching a cluster.
func (o *OCMProvider) IsValidClusterName(clusterName string) (bool, error) {
	// Create a context:
	ctx := context.Background()

	collection := o.conn.ClustersMgmt().V1().Clusters()

	// Retrieve the list of clusters using pages of ten items, till we get a page that has less
	// items than requests, as that marks the end of the collection:
	size := 50
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
// nolint:gocyclo
func (o *OCMProvider) LaunchCluster(clusterName string) (string, error) {
	flavourID := getFlavour()
	skuID := getSKU()
	if skuID != "" {
		// check that enough quota exists for this test if creating cluster
		if enoughQuota, err := o.CheckQuota(skuID); err != nil {
			log.Printf("Failed to check if enough quota is available: %v", err)
		} else if !enoughQuota {
			return "", fmt.Errorf("currently not enough quota exists to run this test")
		}
	} else {
		// If no SKU specified, just continue on with no validation, but log it
		log.Printf("No SKU specified, will not check if enough quota is available.")
	}

	multiAZ := viper.GetBool(config.Cluster.MultiAZ)
	cloudProvider := viper.GetString(config.CloudProvider.CloudProviderID)
	computeMachineType, err := o.DetermineMachineType(cloudProvider)
	if err != nil {
		return "", fmt.Errorf("error while determining machine type: %v", err)
	}

	region, err := o.DetermineRegion(cloudProvider)
	if err != nil {
		return "", fmt.Errorf("error while determining region: %v", err)
	}

	nodeBuilder := &v1.ClusterNodesBuilder{}

	clusterProperties, err := o.GenerateProperties()
	if err != nil {
		return "", fmt.Errorf("error generating cluster properties: %v", err)
	}

	installConfig := ""

	// This skips setting install_config for any prod job OR any periodic addon job.
	// To invoke this logic locally you will have to set JOB_TYPE to "periodic".
	if o.Environment() != "prod" {
		if viper.GetString(config.JobType) == "periodic" && !strings.Contains(viper.GetString(config.JobName), "addon") {
			imageSource := viper.GetString(config.Cluster.ImageContentSource)
			installConfig += "\n" + o.ChooseImageSource(imageSource)
		}
	}

	if viper.GetString(config.Cluster.InstallConfig) != "" {
		installConfig += "\n" + viper.GetString(config.Cluster.InstallConfig)
	}

	if installConfig != "" {
		log.Println("Install config:", installConfig)
		clusterProperties["install_config"] = installConfig
	}

	newCluster := v1.NewCluster().
		Name(clusterName).
		Flavour(v1.NewFlavour().
			ID(flavourID)).
		Region(v1.NewCloudRegion().
			ID(region)).
		MultiAZ(multiAZ).
		Version(v1.NewVersion().
			ID(viper.GetString(config.Cluster.Version)).
			ChannelGroup(viper.GetString(config.Cluster.Channel))).
		CloudProvider(v1.NewCloudProvider().
			ID(cloudProvider)).
		Properties(clusterProperties)

	if viper.GetBool(CCS) {
		// If AWS credentials are set, this must be an AWS CCS cluster
		awsAccount := viper.GetString(AWSAccount)
		awsAccessKey := viper.GetString(AWSAccessKey)
		awsSecretKey := viper.GetString(AWSSecretKey)
		// Refactor: This is a hack to get the AWS CCS cluster to work. In reality today we are loading too many secrets and need a better way to do this.
		// IE: If aws keys are set but not awsAccount, we should mention it's an AWS execution but we are missing credentials.
		if viper.GetString(GCPCredsJSON) != "" {
			gcp, err := v1.UnmarshalGCP(viper.GetString(GCPCredsJSON))
			if err != nil {
				return "", fmt.Errorf("error unmarshalling GCP credentials: %v", err)
			}
			viper.Set(GCPCredsJSON, gcp.Type())
			viper.Set(GCPProjectID, gcp.ProjectID())
			viper.Set(GCPPrivateKeyID, gcp.PrivateKeyID())
			viper.Set(GCPPrivateKey, gcp.PrivateKey())
			viper.Set(GCPClientEmail, gcp.ClientEmail())
			viper.Set(GCPClientID, gcp.ClientID())
			viper.Set(GCPAuthURI, gcp.AuthURI())
			viper.Set(GCPTokenURI, gcp.TokenURI())
			viper.Set(GCPAuthProviderX509CertURL, gcp.AuthProviderX509CertURL())
			viper.Set(gcp.ClientX509CertURL(), gcp.ClientX509CertURL())
		}

		if viper.GetString(config.CloudProvider.CloudProviderID) == "aws" && awsAccount != "" && awsAccessKey != "" && awsSecretKey != "" {
			if viper.GetString(AWSVPCSubnetIDs) != "" {
				subnetIDs := strings.Split(viper.GetString(AWSVPCSubnetIDs), ",")
				awsBuilder := v1.NewAWS().
					AccountID(awsAccount).
					AccessKeyID(awsAccessKey).
					SecretAccessKey(awsSecretKey).
					SubnetIDs(subnetIDs...)
				newCluster = newCluster.CCS(v1.NewCCS().Enabled(true)).AWS(awsBuilder)
				if len(subnetIDs) > 0 {
					cloudProviderData, err := v1.NewCloudProviderData().
						AWS(awsBuilder).
						Region(v1.NewCloudRegion().ID(region)).
						Build()
					if err != nil {
						return "", fmt.Errorf("error building AWS cloud provider data for retrieving Availability Zones: %v", err)
					}
					subnetworks, err := o.GetSubnetworks(cloudProviderData)
					if err != nil {
						return "", fmt.Errorf("error retrieving AWS subnetworks: %v", err)
					}
					availabilityZones := GetAvailabilityZones(subnetworks, subnetIDs)
					nodeBuilder.AvailabilityZones(availabilityZones...)
				}
				if viper.GetBool(config.Cluster.UseProxyForInstall) {
					proxy := v1.NewProxy()
					if userCaBundle := viper.GetString(config.Proxy.UserCABundle); userCaBundle != "" {
						userCaBundleData, err := o.LoadUserCaBundleData(userCaBundle)
						if err != nil {
							return "", fmt.Errorf("error loading CA contents: %v", err)
						}
						newCluster = newCluster.AdditionalTrustBundle(userCaBundleData)
					}
					if httpProxy := viper.GetString(config.Proxy.HttpProxy); httpProxy != "" {
						proxy = proxy.HTTPProxy(httpProxy)
						newCluster = newCluster.Proxy(proxy)
					}
					if httpsProxy := viper.GetString(config.Proxy.HttpsProxy); httpsProxy != "" {
						proxy = proxy.HTTPSProxy(httpsProxy)
						newCluster = newCluster.Proxy(proxy)
					}
				}
			} else {
				_, err = aws.VerifyCCS()
				if err != nil {
					return "", fmt.Errorf("error verifying CCS credentials: %v", err)
				}

				newCluster = newCluster.CCS(v1.NewCCS().Enabled(true)).AWS(
					v1.NewAWS().
						AccountID(awsAccount).
						AccessKeyID(awsAccessKey).
						SecretAccessKey(awsSecretKey))
			}
		} else if viper.GetString(config.CloudProvider.CloudProviderID) == "gcp" && viper.GetString(GCPProjectID) != "" {
			// If GCP credentials are set, this must be a GCP CCS cluster
			newCluster = newCluster.CCS(v1.NewCCS().Enabled(true)).GCP(v1.NewGCP().
				Type(viper.GetString(GCPCredsType)).
				ProjectID(viper.GetString(GCPProjectID)).
				PrivateKey(viper.GetString(GCPPrivateKey)).
				PrivateKeyID(viper.GetString(GCPPrivateKeyID)).
				ClientEmail(viper.GetString(GCPClientEmail)).
				ClientID(viper.GetString(GCPClientID)).
				AuthURI(viper.GetString(GCPAuthURI)).
				TokenURI(viper.GetString(GCPTokenURI)).
				AuthProviderX509CertURL(viper.GetString(GCPAuthProviderX509CertURL)).
				ClientX509CertURL(viper.GetString(GCPClientX509CertURL)))
		} else {
			return "", fmt.Errorf("invalid or no CCS Credentials provided for CCS cluster")
		}
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

	if viper.GetBool(config.Cluster.UseExistingCluster) && viper.GetString(config.Addons.IDs) == "" {
		product := cluster.Product().ID()
		if product == "" {
			product = "osd"
		}
		if clusterID := o.FindRecycledCluster(cluster.Version().ID(), cluster.CloudProvider().ID(), product); clusterID != "" {
			return clusterID, nil
		}
	}

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

func (o *OCMProvider) FindRecycledCluster(originalVersion, cloudProvider, product string) string {
	version := semver.MustParse(strings.TrimPrefix(originalVersion, "openshift-"))
	query := fmt.Sprintf("cloud_provider.id='%s' and properties.JobName='' and properties.JobID='' and product.id='%s' and properties.Status like '%s%%' and version.id like 'openshift-v%s%%'",
		cloudProvider, product, "completed-", version.String())

	log.Println(query)

	listResponse, err := o.conn.ClustersMgmt().V1().Clusters().List().Search(query).Send()
	if err == nil && listResponse.Total() > 0 {
		log.Printf("We've found %d matching clusters to reuse", listResponse.Total())
		recycledCluster := listResponse.Items().Slice()[rand.Intn(listResponse.Total())]
		spiRecycledCluster, err := o.ocmToSPICluster(recycledCluster)
		if err != nil {
			log.Printf("Error converting recycled cluster to an SPI Cluster: %s", err.Error())
			return ""
		}

		if recycledCluster.ExpirationTimestamp().Before(time.Now().Add(4 * time.Hour)) {
			// Let's just expire this cluster immediately
			err = o.AddProperty(spiRecycledCluster, clusterproperties.JobName, "expiring")
			if err != nil {
				log.Printf("Error adding `expiring` to job name: %s", err.Error())
				return ""
			}
			err = o.Expire(spiRecycledCluster.ID())
			if err != nil {
				log.Printf("Error expiring cluster %s: %s", spiRecycledCluster.ID(), err.Error())
				return ""
			}
			// Now try and grab a different existing cluster
			return o.FindRecycledCluster(originalVersion, cloudProvider, product)
		}

		err = o.AddProperty(spiRecycledCluster, clusterproperties.JobID, viper.GetString(config.JobID))
		if err != nil {
			log.Printf("Error adding property to cluster: %s", err.Error())
			return ""
		}
		err = o.AddProperty(spiRecycledCluster, clusterproperties.JobName, viper.GetString(config.JobName))
		if err != nil {
			log.Printf("Error adding property to cluster: %s", err.Error())
			return ""
		}

		time.Sleep(5 * time.Second)

		spiRecycledCluster, err = o.GetCluster(spiRecycledCluster.ID())
		if err != nil {
			log.Printf("Error retrieving cluster during job ID check: %s", err.Error())
			return ""
		}
		if jobID, ok := spiRecycledCluster.Properties()["JobID"]; ok && jobID != viper.GetString(config.JobID) {
			log.Printf("Cluster already recycled by %s", spiRecycledCluster.Properties()["JobID"])
			return o.FindRecycledCluster(originalVersion, cloudProvider, product)
		}

		if recycledCluster.State() == v1.ClusterStateReady {
			err = o.AddProperty(spiRecycledCluster, clusterproperties.Status, clusterproperties.StatusHealthy)
			if err != nil {
				log.Printf("Error adding property to cluster: %s", err.Error())
				return ""
			}
			viper.Set(config.Cluster.Reused, true)
			if recycledCluster.AWS().STS().RoleARN() != "" {
				viper.Set("rosa.STS", true)
			}
			log.Println("Hot cluster ready, moving on...")

			return recycledCluster.ID()
		}
		if recycledCluster.State() == "hibernating" && o.Resume(recycledCluster.ID()) {
			log.Println("Resuming cluster to use...")
			err = o.AddProperty(spiRecycledCluster, clusterproperties.Status, clusterproperties.StatusResuming)
			if err != nil {
				log.Printf("Error adding property to cluster: %s", err.Error())
				return ""
			}
			viper.Set(config.Cluster.Reused, true)
			if recycledCluster.AWS().STS().RoleARN() != "" {
				viper.Set("rosa.STS", true)
			}
			return recycledCluster.ID()
		}
		log.Printf("Failed to recycle cluster %s", recycledCluster.ID())
	}

	// We couldn't resume the cluster. We'll log that we failed but continue the test with a new cluster
	log.Println("Creating a new cluster instead")
	return ""
}

// DetermineRegion will return the region provided by configs. This mainly wraps the random functionality for use
// by the ROSA provider.
func (o *OCMProvider) DetermineRegion(cloudProvider string) (string, error) {
	region := viper.GetString(config.CloudProvider.Region)

	// If a region is set to "random", it will poll OCM for all the regions available
	// It then will pull a random entry from the list of regions and set the ID to that
	if region == "random" {
		var regions []*v1.CloudRegion
		// We support multiple cloud providers....
		if cloudProvider == "aws" {
			if viper.GetString(AWSAccessKey) == "" || viper.GetString(AWSSecretKey) == "" {
				log.Println("Random region requested but cloud credentials not supplied. Defaulting to us-east-1")
				return "us-east-1", nil
			}
			awsCredentials, err := v1.NewAWS().
				AccessKeyID(viper.GetString(AWSAccessKey)).
				SecretAccessKey(viper.GetString(AWSSecretKey)).
				Build()
			if err != nil {
				return "", err
			}

			response, err := o.conn.ClustersMgmt().V1().CloudProviders().CloudProvider(cloudProvider).AvailableRegions().Search().Body(awsCredentials).Send()
			if err != nil {
				log.Printf("Error selecting region: %s", err.Error())
				log.Println("Defaulting to us-east-1")
				region := "us-east-1"
				viper.Set(config.CloudProvider.Region, region)
				return region, nil
			}
			regions = response.Items().Slice()
		}

		// But we don't support passing GCP credentials yet :)
		if cloudProvider == "gcp" {
			log.Println("Random GCP region not supported yet. Setting region to us-east1")
			return "us-east1", nil
		}

		cloudRegion, found := ChooseRandomRegion(toCloudRegions(regions)...)
		if !found {
			return "", fmt.Errorf("unable to choose a random enabled region")
		}

		region = cloudRegion.ID()

		log.Printf("Random region requested, selected %s region.", region)

		// Update the Config with the selected random region
		viper.Set(config.CloudProvider.Region, region)
	}

	return region, nil
}

// CloudRegion provides an interface for methods on *v1.CloudRegion so that
// compatible types can be instantiated from tests.
type CloudRegion interface {
	ID() string
	Enabled() bool
}

// ensure *v1.CloudRegion implements CloudRegion at compile time
var _ CloudRegion = &v1.CloudRegion{}

// toCloudRegions converts a slice of *v1.CloudRegion into a slice of CloudRegion.
// This helper can be removed once generics lands in Go, as this will no longer be
// necessary.
func toCloudRegions(in []*v1.CloudRegion) []CloudRegion {
	out := make([]CloudRegion, 0, len(in))
	for i := range in {
		out = append(out, in[i])
	}
	return out
}

// ChooseRandomRegion chooses a random enabled region from the provided options. Its
// second return parameter indicates whether it was successful in finding an enabled
// region.
func ChooseRandomRegion(regions ...CloudRegion) (CloudRegion, bool) {
	// remove disabled regions from consideration
	enabledRegions := make([]CloudRegion, 0, len(regions))
	for _, region := range regions {
		if region.Enabled() {
			enabledRegions = append(enabledRegions, region)
		}
	}
	// randomize the order of the candidates
	rand.Shuffle(len(enabledRegions), func(i, j int) {
		enabledRegions[i], enabledRegions[j] = enabledRegions[j], enabledRegions[i]
	})
	// return the first element if the list is not empty
	for _, regionObj := range enabledRegions {
		return regionObj, true
	}
	// indicate that there were no enabled candidates
	return nil, false
}

// DetermineMachineType will return the machine type provided by configs. This mainly wraps the random functionality for use by the OCM provider.
// Returns a random machine type if the machine type is set to "random" and a more narrowed random if a regex was specified.
func (o *OCMProvider) DetermineMachineType(cloudProvider string) (string, error) {
	computeMachineType, computeMachineTypeRegex := viper.GetString(ComputeMachineType), viper.GetString(ComputeMachineTypeRegex)
	searchString, returnedType := "", ""

	// If a machineType is set to "random", it will poll OCM for all the machines available
	// It then will pull a random entry from the list of machines and set the machineTypes to that
	if (computeMachineType == "random" && rand.Intn(3) >= 2) || (computeMachineType == "random" && computeMachineTypeRegex != "") {
		// Create search string based on wether we are using a regex or not
		switch {
		case computeMachineType == "random" && computeMachineTypeRegex != "":
			searchString = fmt.Sprintf("cloud_provider.id like '%s' AND id like '%s.%%'", cloudProvider, computeMachineTypeRegex)
		case computeMachineType == "random" && computeMachineTypeRegex == "":
			searchString = fmt.Sprintf("cloud_provider.id like '%s'", cloudProvider)
		}
		machinetypeClient := o.conn.ClustersMgmt().V1().MachineTypes().List().Search(searchString)
		log.Printf("Randomly picking size for MachineTypes with search string %s", computeMachineTypeRegex)

		machinetypes, err := machinetypeClient.Send()
		if err != nil {
			return "", err
		}

		for range machinetypes.Items().Slice() {
			machinetypeObj := machinetypes.Items().Slice()[rand.Intn(machinetypes.Total())]
			returnedType = machinetypeObj.ID()
			break
		}
		log.Printf("Random machine type requested, selected `%s` machine type.", returnedType)
	}

	if computeMachineType != "random" && computeMachineType != "" {
		log.Printf("Machine type manually set to %s", computeMachineType)
		returnedType = computeMachineType
	}

	// Update the Config with the selected random machine
	viper.Set(ComputeMachineType, returnedType)

	return returnedType, nil
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
		clusterproperties.JobName:          viper.GetString(config.JobName),
		clusterproperties.JobID:            viper.GetString(config.JobID),
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
	var deleteResp *v1.ClusterDeleteResponse
	var resumeResp *v1.ClusterResumeResponse
	var cluster *spi.Cluster
	var err error

	cluster, err = o.GetCluster(clusterID)
	if err != nil {
		return fmt.Errorf("error retrieving cluster for deletion: %v", err)
	}

	// If the cluster is hibernating according to OCM, wake it up
	if cluster.State() == spi.ClusterStateHibernating {
		if err = retryer().Do(func() error {
			var err error
			resumeResp, err = o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).Resume().Send()

			if err != nil {
				return fmt.Errorf("couldn't resume cluster '%s': %v", clusterID, err)
			}

			if resumeResp != nil && resumeResp.Error() != nil {
				err = errResp(resumeResp.Error())
				log.Printf("%v", err)
				return err
			}

			return nil
		}); err != nil {
			return err
		}
	}

	wait.PollImmediate(1*time.Minute, 15*time.Minute, func() (bool, error) {
		// If the cluster state is anything but Hibernating or Ready, poll the state again
		if cluster.State() == spi.ClusterStateHibernating || cluster.State() == spi.ClusterStateReady {
			cluster, err = o.GetCluster(clusterID)
			if err != nil {
				log.Printf("error retrieving cluster for deletion: %v", err)
				return false, nil
			}
		}
		// A cluster errored in OCM is unlikely to recover so we should fail fast
		if cluster.State() == spi.ClusterStateError {
			return false, fmt.Errorf("Cluster %s is in an errored state.", cluster.ID())
		}

		// We have a ready cluster, hooray
		if cluster.State() == spi.ClusterStateReady {
			return true, nil
		}

		// The cluster isn't ready so we should loop again
		return false, nil
	})

	err = o.AddProperty(cluster, clusterproperties.Status, clusterproperties.StatusUninstalling)
	if err != nil {
		return fmt.Errorf("error adding uninstalling status to cluster: %v", err)
	}

	err = retryer().Do(func() error {
		var err error
		deleteResp, err = o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).
			Delete().
			Send()

		if err != nil {
			return fmt.Errorf("couldn't delete cluster '%s': %v", clusterID, err)
		}

		if deleteResp != nil && deleteResp.Error() != nil {
			err = errResp(deleteResp.Error())
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

	for len(clusters) < totalItems || !emptyPage {
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
	// Override with a local kubeconfig if defined
	localKubeConfig := viper.GetString(config.Kubeconfig.Path)
	if len(localKubeConfig) > 0 {
		log.Printf("Overriding provider kubeconfig with local: %s", localKubeConfig)
		return getLocalKubeConfig(localKubeConfig)
	}

	existingKubeConfig := viper.GetString(config.Kubeconfig.Contents)
	if existingKubeConfig != "" {
		return []byte(existingKubeConfig), nil
	}

	if creds, ok := o.credentialCache[clusterID]; ok {
		viper.Set(config.Kubeconfig.Contents, creds)
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
	viper.Set(config.Kubeconfig.Contents, resp.Body().Kubeconfig())

	return []byte(resp.Body().Kubeconfig()), nil
}

// Loads and returns the supplied filepath
func getLocalKubeConfig(path string) ([]byte, error) {
	fileReader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	f, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return nil, err
	}
	return []byte(f), nil
}

// InstallAddons loops through the addons list in the config
// and performs the CRUD operation to trigger addon installation
func (o *OCMProvider) InstallAddons(clusterID string, addonIDs []spi.AddOnID, addonParams map[spi.AddOnID]spi.AddOnParams) (num int, err error) {
	num = 0
	addonsClient := o.conn.ClustersMgmt().V1().Addons()
	clusterClient := o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID)
	for _, addonID := range addonIDs {
		var addonResp *v1.AddOnGetResponse
		params := addonParams[addonID]

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
			builder := v1.NewAddOnInstallation().Addon(v1.NewAddOn().Copy(addon))

			if len(params) > 0 {
				addOnParamList := make([]*v1.AddOnInstallationParameterBuilder, 0, len(params))
				for name, value := range params {
					addOnParamList = append(addOnParamList, v1.NewAddOnInstallationParameter().ID(name).Value(value))
				}
				builder = builder.Parameters(v1.NewAddOnInstallationParameterList().Items(addOnParamList...))
			}

			addonInstallation, err := builder.Build()
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

	if product, ok := ocmCluster.GetProduct(); ok {
		cluster.Product(product.ID())
	}

	if state, ok := ocmCluster.GetState(); ok {
		cluster.State(ocmStateToInternalState(state))
	}

	if properties, ok := ocmCluster.GetProperties(); ok {
		cluster.Properties(properties)
	}

	if !viper.GetBool(config.Addons.SkipAddonList) {
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

	}

	cluster.ExpirationTimestamp(ocmCluster.ExpirationTimestamp())
	cluster.CreationTimestamp(ocmCluster.CreationTimestamp())
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
	case v1.ClusterStateHibernating:
		return spi.ClusterStateHibernating
	case v1.ClusterStateResuming:
		return spi.ClusterStateResuming
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

	if extendexpirytime.Year() < 2000 {
		log.Println("Cluster does not have an expiration!")
		return nil
	}

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

// Expire sets the expiration time of an existing cluster to the current time
func (o *OCMProvider) Expire(clusterID string) error {
	var resp *v1.ClusterUpdateResponse

	now := time.Now().Add(1 * time.Minute)

	extendexpiryCluster, err := v1.NewCluster().ExpirationTimestamp(now).Build()
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

	log.Println("Successfully set cluster expiry time to", now.UTC().Format("Monday, 02 Jan 2006 15:04:05 MST"))

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

	if viper.GetString(config.JobName) != "" {
		propertyFilename := fmt.Sprintf("%s.osde2e-cluster-property-update.metrics.prom", cluster.ID())
		data := fmt.Sprintf("# TYPE cicd_cluster_properties gauge\ncicd_cluster_properties{cluster_id=\"%s\",environment=\"%s\",job_id=\"%s\",property=\"%s\",region=\"%s\",value=\"%s\",version=\"%s\"} 0\n", cluster.ID(), o.Environment(), viper.GetString(config.JobID), tag, cluster.Region(), value, cluster.Version())
		err = aws.WriteToS3(aws.CreateS3URL(viper.GetString(config.Tests.MetricsBucket), "incoming", propertyFilename), []byte(data))
		if err != nil {
			return fmt.Errorf("failed to upload cluster property metrics: %v", err)
		}
		log.Println(data)
	}

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
func (o *OCMProvider) Upgrade(clusterID string, version string, t time.Time) error {
	policy, err := v1.NewUpgradePolicy().Version(version).NextRun(t).ScheduleType("manual").Build()
	if err != nil {
		return err
	}

	addResp, err := o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).UpgradePolicies().Add().Body(policy).SendContext(context.TODO())
	if err != nil {
		return err
	}
	if addResp.Status() != http.StatusCreated {
		log.Printf("Unable to schedule upgrade with provider (status %d, response %v)", addResp.Status(), addResp.Error())
		return addResp.Error()
	}

	log.Printf("upgrade to version %s scheduled with provider for time %s", addResp.Body().Version(), addResp.Body().NextRun().Format(time.RFC3339))
	return nil
}

// GetUpgradePolicyID gets the first upgrade policy from the top
func (o *OCMProvider) GetUpgradePolicyID(clusterID string) (string, error) {
	listResp, err := o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).UpgradePolicies().List().SendContext(context.TODO())
	if err != nil {
		return "", err
	}
	if listResp.Status() != http.StatusOK {
		log.Printf("Unable to find upgrade schedule with provider (status %d, response %v)", listResp.Status(), listResp.Error())
		return "", listResp.Error()
	}

	if listResp.Items().Len() < 1 {
		// Don't treat this as an error (because it may not be), just return nothing
		log.Printf("No upgrade policies currently exist on the provider.")
		return "", nil
	}

	policyID := listResp.Items().Get(0).ID()
	if policyID == "" {
		return "", fmt.Errorf("failed to get the policy ID")
	}

	return policyID, nil
}

// UpdateSchedule updates the existing upgrade policy for re-scheduling
func (o *OCMProvider) UpdateSchedule(clusterID string, version string, t time.Time, policyID string) error {
	policyBody, err := v1.NewUpgradePolicy().NextRun(t).Build()
	if err != nil {
		return err
	}

	updateResp, err := o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).UpgradePolicies().UpgradePolicy(policyID).Update().Body(policyBody).SendContext(context.TODO())
	if err != nil {
		return err
	}
	if updateResp.Status() != http.StatusOK {
		log.Printf("Unable to update upgrade schedule with provider (status %d, response %v)", updateResp.Status(), updateResp.Error())
		return err
	}

	log.Printf("Update the upgrade schedule for cluster %s to %s", clusterID, t)
	return nil
}

// Resume resumes a cluster via OCM
func (o *OCMProvider) Resume(id string) bool {
	resp, err := o.conn.ClustersMgmt().V1().Clusters().Cluster(id).Resume().Send()
	if err != nil {
		err = fmt.Errorf("couldn't resume cluster '%s': %v", id, err)
		log.Printf("%v", err)
		return false
	}

	if resp != nil && resp.Error() != nil {
		log.Printf("error while trying to resume cluster: %v", err)
		return false
	}

	return true
}

// Hibernate resumes a cluster via OCM
func (o *OCMProvider) Hibernate(id string) bool {
	resp, err := o.conn.ClustersMgmt().V1().Clusters().Cluster(id).Hibernate().Send()
	if err != nil {
		err = fmt.Errorf("couldn't hibernate cluster '%s': %v", id, err)
		log.Printf("%v", err)
		return false
	}

	if resp != nil && resp.Error() != nil {
		log.Printf("error while trying to hibernate cluster: %v", err)
		return false
	}

	return true
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

func (o *OCMProvider) GetSubnetworks(cloudProviderData *v1.CloudProviderData) (subnetworks []*v1.Subnetwork, err error) {
	if viper.GetBool(CCS) && viper.GetString(config.CloudProvider.CloudProviderID) == "aws" {
		response, err := o.conn.ClustersMgmt().V1().AWSInquiries().Vpcs().Search().
			Page(1).
			Size(-1).
			Body(cloudProviderData).
			Send()
		if err != nil {
			return nil, err
		}

		cloudVPCs := response.Items().Slice()

		for _, vpc := range cloudVPCs {
			subnetworks = append(subnetworks, vpc.AWSSubnets()...)
		}
	}
	return subnetworks, nil
}

func GetAvailabilityZones(subnetworks []*v1.Subnetwork, configSubnetIDs []string) (availabilityZones []string) {
	collectedAZs := map[string]bool{}
	for _, subnet := range subnetworks {
		subnetID := subnet.SubnetID()
		availabilityZone := subnet.AvailabilityZone()
		for _, configSubnetID := range configSubnetIDs {
			if subnetID != configSubnetID || collectedAZs[availabilityZone] {
				continue
			}
			collectedAZs[availabilityZone] = true
			availabilityZones = append(availabilityZones, availabilityZone)
		}
	}
	return availabilityZones
}
