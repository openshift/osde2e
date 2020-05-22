package ocmprovider

import (
	"fmt"
	"log"
	"os/user"
	"time"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/openshift/osde2e/pkg/common/state"
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
	cfg := config.Instance
	state := state.Instance

	log.Printf("Creating cluster '%s'...", state.Cluster.Name)

	// choose flavour based on config
	flavourID := DefaultFlavour

	// Calculate an expiration date for the cluster so that it will be automatically deleted if
	// we happen to forget to do it:
	expiration := time.Now().Add(time.Duration(cfg.Cluster.ExpiryInMinutes) * time.Minute).UTC() // UTC() to workaround SDA-1567.

	var username string

	// If JobID is not equal to -1, then we're running on prow.
	if cfg.JobID != -1 {
		username = "prow"
	} else {

		user, err := user.Current()

		if err != nil {
			return "", fmt.Errorf("unable to get current user: %v", err)
		}

		username = user.Username
	}

	newCluster := v1.NewCluster().
		Name(state.Cluster.Name).
		Flavour(v1.NewFlavour().
			ID(flavourID)).
		Region(v1.NewCloudRegion().
			ID(state.CloudProvider.Region)).
		MultiAZ(cfg.Cluster.MultiAZ).
		Version(v1.NewVersion().
			ID(state.Cluster.Version)).
		CloudProvider(v1.NewCloudProvider().
			ID(state.CloudProvider.CloudProviderID)).
		ExpirationTimestamp(expiration).
		Properties(map[string]string{
			MadeByOSDe2e: "true",
			OwnedBy:      username,
		})

	// Configure the cluster to be Multi-AZ if configured
	// We must manually configure the number of compute nodes
	// Currently set to 9 nodes. Whatever it is, must be divisible by 3.
	if cfg.Cluster.MultiAZ {
		numNodes := &v1.ClusterNodesBuilder{}

		newCluster = newCluster.
			Nodes(numNodes.Compute(9)).
			MultiAZ(cfg.Cluster.MultiAZ)
	}

	if len(cfg.Addons.IDsAtCreation) > 0 {
		addons := []*v1.AddOnInstallationBuilder{}
		for _, id := range cfg.Addons.IDsAtCreation {
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

// GetCluster returns a cluster from OCM.
func (o *OCMProvider) GetCluster(clusterID string) (*spi.Cluster, error) {
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

	ocmCluster := resp.Body()

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

	var addonsResp *v1.AddOnInstallationsListResponse
	err = retryer().Do(func() error {
		var err error
		addonsResp, err = o.conn.ClustersMgmt().V1().Clusters().Cluster(clusterID).Addons().
			List().
			Send()

		if err != nil {
			err = fmt.Errorf("couldn't retrieve addons for cluster '%s': %v", clusterID, err)
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

	return cluster.Build(), nil
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
