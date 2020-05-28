package query

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"time"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/common/prometheus"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

var Cmd = &cobra.Command{
	Use:   "query",
	Short: "Queries Prometheus results.",
	Long:  "Queries Prometheus results.",
	Args:  cobra.OnlyValidArgs,
	RunE:  run,
}

var args struct {
	configString string
	customConfig string
	outputFormat string
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
		&args.outputFormat,
		"output-format",
		"-",
		"Output format for query results (json|prom). Defaults to json.",
	)

	Cmd.RegisterFlagCompletionFunc("output-format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "prom"}, cobra.ShellCompDirectiveDefault
	})
}

func run(cmd *cobra.Command, argv []string) error {

	if err := common.LoadConfigs(args.configString, args.customConfig); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	query := strings.Join(argv, " ")

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

	return nil
}
