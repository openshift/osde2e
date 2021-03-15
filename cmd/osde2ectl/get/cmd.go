package get

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/common/clusterproperties"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "get",
	Short: "Get information on created/existing clusters made by osde2e",
	Long:  "Get information about a CI cluster or its kubeconfig using arguments.",
	Args:  cobra.OnlyValidArgs,
	RunE:  run,
}

var provider spi.Provider

var args struct {
	configString    string
	customConfig    string
	secretLocations string
	clusterID       string
	kubeConfig      bool
	kubeConfigPath  string
	hours           uint64
	minutes         uint64
}

func init() {
	pfs := Cmd.PersistentFlags()

	pfs.StringVar(
		&args.configString,
		"configs",
		"",
		"A comma separated list of built in configs to use",
	)
	pfs.StringVar(
		&args.customConfig,
		"custom-config",
		"",
		"Custom config file for osde2e",
	)
	pfs.StringVar(
		&args.secretLocations,
		"secret-locations",
		"",
		"A comma separated list of possible secret directory locations for loading secret configs.",
	)
	pfs.StringVar(
		&args.kubeConfigPath,
		"kube-config-path",
		"",
		"Path to place the downloaded kubeconfig info about a cluster",
	)
	pfs.StringVarP(
		&args.clusterID,
		"cluster-id",
		"i",
		"",
		"Existing OCM cluster ID to get information about the cluster.",
	)
	pfs.BoolVarP(
		&args.kubeConfig,
		"kube-config",
		"k",
		false,
		"A flag that triggers the fetching of a given cluster's kubeconfig.",
	)
	pfs.Uint64VarP(
		&args.hours,
		"hours",
		"r",
		0,
		"Cluster expiration extension value in hours.",
	)

	pfs.Uint64VarP(
		&args.minutes,
		"minutes",
		"m",
		0,
		"Cluster expiration extension value in minutes.",
	)

}

func run(cmd *cobra.Command, argv []string) error {

	if err := common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	kubeconfigStatus, err := cmd.PersistentFlags().GetBool("kube-config")

	if err != nil {
		return fmt.Errorf("error retrieving kube-config information: %v", err)
	}

	if provider, err = providers.ClusterProvider(); err != nil {
		return fmt.Errorf("could not setup cluster provider: %v", err)
	}

	clusterID := args.clusterID

	cluster, err := provider.GetCluster(clusterID)

	if err != nil {
		return fmt.Errorf("error getting cluster: %v", err)
	}

	timediff := cluster.ExpirationTimestamp().UTC().Sub(time.Now().UTC()).Minutes()

	if err != nil {
		return fmt.Errorf("error retrieving cluster information: %v", err)
	}

	properties := cluster.Properties()
	if properties[clusterproperties.MadeByOSDe2e] != "true" {
		return fmt.Errorf("Cluster was not created by osde2e")
	}

	fmt.Printf("%17s: %s\n", "NAME", cluster.Name())
	fmt.Printf("%17s: %s\n", "ID", cluster.ID())
	fmt.Printf("%17s: %s\n", "STATE", cluster.State())
	fmt.Printf("%17s: %s\n", "STATUS", properties[clusterproperties.Status])
	fmt.Printf("%17s: %s\n", "OWNER", properties[clusterproperties.OwnedBy])
	fmt.Printf("%17s: %s\n", "INSTALLED VERSION", properties[clusterproperties.InstalledVersion])
	fmt.Printf("%17s: %s\n", "UPGRADE VERSION", properties[clusterproperties.UpgradeVersion])

	if jobName, ok := properties[clusterproperties.JobName]; ok {
		fmt.Printf("%17s: %s\n", "JOB NAME", jobName)
	}

	if jobID, ok := properties[clusterproperties.JobID]; ok {
		fmt.Printf("%17s: %s\n", "JOB ID", jobID)
	}

	if kubeconfigStatus {
		content, err := getKubeconfig(clusterID)
		if err != nil {
			return fmt.Errorf("Error getting the cluster's kubeconfig - %s", err)
		}

		filename := cluster.Name() + "-kubeconfig.txt"

		var filePath string
		if args.kubeConfigPath != "" {
			_, err := os.Stat(args.kubeConfigPath)
			if os.IsNotExist(err) {
				fmt.Println("File directory is invalid - ", err.Error(), "Will create a new directory.")
				err = os.Mkdir(args.kubeConfigPath, os.ModePerm)
				if err != nil {
					return fmt.Errorf("Unable to create a new directory - %s", err.Error())
				}
			}
			if err != nil {
				return fmt.Errorf("Unable to stat file path: %s", err.Error())
			}
			filePath = filepath.Join(args.kubeConfigPath, filename)
		} else {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("Unable to get CWD: %s", err.Error())
			}

			filePath = filepath.Join(string(dir), filename)
		}
		err = ioutil.WriteFile(filePath, content, os.ModePerm)

		if err != nil {
			return fmt.Errorf("could not write KubeConfig into a file: %v", err)
		}
		fmt.Println("Successfully downloaded the kubeconfig into " + filePath + ". Run the command:\n\nexport TEST_KUBECONFIG=\"" + filePath + "\"\n")
		if args.hours == 0 && args.minutes == 0 {
			if timediff <= 30 {
				fmt.Println("Cluster expiry time is less than 30 minutes. Extending expiry time by 30 minutes.")
				args.minutes = 30
				if err = provider.ExtendExpiry(clusterID, 0, 30, 0); err != nil {
					return fmt.Errorf("error extending cluster expiry time: %s", err.Error())
				}
				fmt.Println("Extended cluster expiry time by :", args.hours, "h ", args.minutes, "m")
			}
		} else {
			if err = provider.ExtendExpiry(clusterID, args.hours, args.minutes, 0); err != nil {
				return fmt.Errorf("error extending cluster expiry time: %s", err.Error())
			}
			fmt.Println("Extended cluster expiry time by :", args.hours, "h ", args.minutes, "m")
		}
	}

	return nil
}

func getKubeconfig(clusterID string) ([]byte, error) {
	provider, err := providers.ClusterProvider()
	var kubeconfigBytes []byte
	if err != nil {
		return kubeconfigBytes, fmt.Errorf("could not setup cluster provider: %v", err)
	}

	if kubeconfigBytes, err = provider.ClusterKubeconfig(clusterID); err != nil {
		return kubeconfigBytes, fmt.Errorf("could not get kubeconfig for cluster: %v", err)
	}
	return kubeconfigBytes, nil
}
