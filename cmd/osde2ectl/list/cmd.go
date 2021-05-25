package list

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/common/clusterproperties"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List created/existing clusters made by osde2e",
	Long:  "List specific clusters using the provided arguments.",
	Args:  cobra.OnlyValidArgs,
	RunE:  run,
}

var args struct {
	configString    string
	customConfig    string
	secretLocations string
	clusterStatus   string
	version         string
	ownedBy         string
}

func init() {
	flags := Cmd.Flags()

	flags.StringVar(
		&args.configString,
		"configs",
		"",
		"A comma separated list of built in configs to use",
	)
	flags.StringVar(
		&args.customConfig,
		"custom-config",
		"",
		"Custom config file for osde2e",
	)
	flags.StringVar(
		&args.secretLocations,
		"secret-locations",
		"",
		"A comma separated list of possible secret directory locations for loading secret configs.",
	)
	flags.StringVarP(
		&args.clusterStatus,
		"cluster-status",
		"s",
		"",
		"A flag to indicate the cluster status parameter in the cluster list query",
	)
	flags.StringVarP(
		&args.version,
		"version",
		"i",
		"",
		"A flag to indicate the cluster version parameter in the cluster list query",
	)
	flags.StringVarP(
		&args.ownedBy,
		"cluster-owner",
		"o",
		"",
		"A flag to indicate the cluster owner parameter in the cluster list query",
	)

	Cmd.RegisterFlagCompletionFunc("output-format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "prom"}, cobra.ShellCompDirectiveDefault
	})
}

func run(cmd *cobra.Command, argv []string) error {
	var provider spi.Provider
	var err error
	if err = common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	if provider, err = providers.ClusterProvider(); err != nil {
		return fmt.Errorf("could not setup cluster provider: %v", err)
	}

	querystring := "properties.MadeByOSDe2e='true'"

	propertymap := map[string]string{
		args.clusterStatus: "properties.Status",
		args.ownedBy:       "properties.OwnedBy",
		args.version:       "version.id",
	}

	filterlist := []string{
		args.ownedBy, args.clusterStatus, args.version,
	}

	for _, filter := range filterlist {
		if filter != "" {
			querystring += fmt.Sprintf(" and %s like '%s'", propertymap[filter], "%%"+filter+"%%")
		}
	}

	metadata.Instance.SetEnvironment(provider.Environment())

	clusters, err := provider.ListClusters(querystring)
	if err != nil {
		return err
	}

	if len(clusters) == 0 {
		log.Printf("No results found")
	} else {
		statuslengthmax, versionlengthmax, ownerlengthmax := 0, 0, 0
		for _, cluster := range clusters {
			properties := cluster.Properties()
			if len(properties[clusterproperties.Status]) > statuslengthmax {
				statuslengthmax = len(properties[clusterproperties.Status])
			}
			if len(cluster.Version()) > versionlengthmax {
				versionlengthmax = len(cluster.Version())
			}
			if len(properties[clusterproperties.OwnedBy]) > ownerlengthmax {
				ownerlengthmax = len(properties[clusterproperties.OwnedBy])
			}
		}

		headerspace := "%-15s%-" + strconv.Itoa(ownerlengthmax+5) + "s%-35s%-10s%-15s%-" + strconv.Itoa(statuslengthmax+5) + "s%-" + strconv.Itoa(versionlengthmax-5) + "s%-25s%s\n"
		fmt.Printf(headerspace, "NAME", "OWNER", "ID", "CLOUD", "STATE", "STATUS", "VERSION", "JOB_ID", "JOB")
		for _, cluster := range clusters {
			properties := cluster.Properties()
			fmt.Printf(headerspace, cluster.Name(), properties[clusterproperties.OwnedBy], cluster.ID(), fmt.Sprintf("%s/%s", cluster.Product(), cluster.CloudProvider()), cluster.State(), properties[clusterproperties.Status], strings.TrimLeft(cluster.Version(), "openshift-v"), properties[clusterproperties.JobID], properties[clusterproperties.JobName])
		}
	}

	return nil
}
