package query

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/spf13/cobra"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/prometheus"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/spi"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prometheusmodel "github.com/prometheus/common/model"
)

var Cmd = &cobra.Command{
	Use:   "query",
	Short: "Queries Prometheus results.",
	Long:  "Queries Prometheus results.",
	Args:  cobra.ArbitraryArgs,
	RunE:  run,
}

var args struct {
	configString    string
	customConfig    string
	secretLocations string
	outputFormat    string
	versionCheck    bool
	installVersion  string
	upgradeVersion  string
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
	flags.StringVar(
		&args.outputFormat,
		"output-format",
		"-",
		"Output format for query results (json|prom). Defaults to json.",
	)
	flags.BoolVarP(
		&args.versionCheck,
		"version-check",
		"v",
		false,
		"A flag that triggers a query that lists valid installed/upgrade versions",
	)
	flags.StringVarP(
		&args.installVersion,
		"install-version",
		"i",
		"",
		"The cluster install version input for version query",
	)
	flags.StringVarP(
		&args.upgradeVersion,
		"upgrade-version",
		"u",
		"",
		"The cluster upgrade version input for upgrade query",
	)

	Cmd.RegisterFlagCompletionFunc("output-format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "prom"}, cobra.ShellCompDirectiveDefault
	})
}

func run(cmd *cobra.Command, argv []string) error {
	if err := common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	var query string

	var count prometheusmodel.Sample

	if !args.versionCheck {
		query = strings.Join(argv, " ")
	} else {
		var provider spi.Provider
		var err error
		cloudprovider := viper.GetString(config.CloudProvider.CloudProviderID)
		if provider, err = providers.ClusterProvider(); err != nil {
			return fmt.Errorf("could not setup cluster provider: %v", err)
		}
		environment := provider.Environment()
		if args.upgradeVersion == "" {
			query = fmt.Sprintf("count by (install_version) (cicd_jUnitResult{cloud_provider=\"%s\",install_version=\"%s\", environment=\"%s\", result=~\"passed|skipped\"}) / count by (install_version) (cicd_jUnitResult{cloud_provider=\"%s\",install_version=\"%s\", environment=\"%s\"})",
				escapeQuotes(cloudprovider), escapeQuotes(args.installVersion), escapeQuotes(environment),
				escapeQuotes(cloudprovider), escapeQuotes(args.installVersion), escapeQuotes(environment))
		} else {
			query = fmt.Sprintf("count by (install_version, upgrade_version) (cicd_jUnitResult{cloud_provider=\"%s\",install_version=\"%s\", upgrade_version=\"%s\", environment=\"%s\", result=~\"passed|skipped\"}) / count by (install_version) (cicd_jUnitResult{cloud_provider=\"%s\",install_version=\"%s\", upgrade_version=\"%s\", environment=\"%s\"})",
				escapeQuotes(cloudprovider), escapeQuotes(args.installVersion), escapeQuotes(args.upgradeVersion), escapeQuotes(environment),
				escapeQuotes(cloudprovider), escapeQuotes(args.installVersion), escapeQuotes(args.upgradeVersion), escapeQuotes(environment))
		}
	}

	client, err := prometheus.CreateClient()
	if err != nil {
		return fmt.Errorf("unable to create Prometheus client: %v", err)
	}

	promAPI := v1.NewAPI(client)
	context, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	value, warnings, err := promAPI.Query(context, query, time.Now())
	if err != nil {
		return fmt.Errorf("error issuing query: %v", err)
	}

	for _, warning := range warnings {
		log.Printf("warning: %s", warning)
	}

	var data []byte
	switch args.outputFormat {
	case "json":
		data, err = json.MarshalIndent(value, "", "  ")

		if err != nil {
			return fmt.Errorf("error marshaling results: %v", err)
		}
	case "prom":
		data = []byte(value.String())
	}

	_, err = os.Stdout.Write(data)

	if err != nil {
		return fmt.Errorf("error writing output: %v", err)
	}

	if args.versionCheck {
		data, err = json.MarshalIndent(value, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling results: %v", err)
		}
		encoded := []rune(string(data))
		data = []byte(string(encoded[1:(len(string(data)) - 1)]))
		count.UnmarshalJSON(data)
		log.Printf("Valid version check count ratio - %v", count.Value)
	}

	return nil
}

func escapeQuotes(stringToEscape string) string {
	return strings.Replace(stringToEscape, `"`, `\"`, -1)
}
