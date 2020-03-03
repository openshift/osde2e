package common

import (
	"fmt"
	"log"
	"strings"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/load"
	"github.com/openshift/osde2e/pkg/common/state"
)

// LoadConfigs loads config objects given the provided list of configs and a custom config
func LoadConfigs(configString string, customConfig string) error {
	var configs []string
	if configString != "" {
		configs = strings.Split(configString, ",")
	}

	for _, config := range configs {
		log.Printf("Will load config %s", config)
	}

	// Load config and initial state
	if err := load.IntoObject(config.Instance, configs, customConfig); err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	if err := load.IntoObject(state.Instance, configs, customConfig); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	return nil
}
