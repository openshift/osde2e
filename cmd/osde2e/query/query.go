package query

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/subcommands"
	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/pkg/common/prometheus"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// Command is the command for querying Prometheus results
type Command struct {
	configString string
	customConfig string

	outputFormat string

	subcommands.Command
}

// Name is the name of the query command
func (*Command) Name() string {
	return "query"
}

// Synopsis is a short summary of the query command
func (*Command) Synopsis() string {
	return "Queries Prometheus results."
}

// Usage describes how the query command is used
func (*Command) Usage() string {
	return "query <query>"
}

// SetFlags describes the arguments used by the query command
func (t *Command) SetFlags(f *flag.FlagSet) {
	f.StringVar(&t.configString, "configs", "", "A comma separated list of built in configs to use")
	f.StringVar(&t.customConfig, "custom-config", "", "Custom config file for osde2e")
	f.StringVar(&t.outputFormat, "output-format", "json", "Custom config file for osde2e")
}

// Execute actually executes the tests
func (t *Command) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() == 0 {
		log.Printf("Unexpected number of arguments.")
		log.Printf(t.Usage())
		return subcommands.ExitFailure
	}

	if err := common.LoadConfigs(t.configString, t.customConfig); err != nil {
		log.Printf("error loading initial state: %v", err)
		return subcommands.ExitFailure
	}

	query := strings.Join(f.Args(), " ")

	client, err := prometheus.CreateClient()

	if err != nil {
		log.Printf("unable to create Prometheus client: %v", err)
		return subcommands.ExitFailure
	}

	promAPI := v1.NewAPI(client)
	context, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	value, warnings, err := promAPI.Query(context, query, time.Now())

	if err != nil {
		log.Printf("error issuing query: %v", err)
		return subcommands.ExitFailure
	}

	for _, warning := range warnings {
		log.Printf("warning: %s", warning)
	}

	var data []byte
	switch t.outputFormat {
	case "json":
		data, err = json.MarshalIndent(value, "", "  ")

		if err != nil {
			log.Printf("error marshaling results: %v", err)
			return subcommands.ExitFailure
		}
	case "prom":
		data = []byte(value.String())
	}

	_, err = os.Stdout.Write(data)

	if err != nil {
		log.Printf("error writing output: %v", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
