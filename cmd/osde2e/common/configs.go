package common

import (
	"fmt"
	"log"
	"strings"

	"github.com/openshift/osde2e/pkg/common/load"
)

// LoadConfigs loads config objects given the provided list of configs and a custom config
func LoadConfigs(configString string, customConfig string, secretLocationsString string) error {
	var configs []string
	if configString != "" {
		configs = strings.Split(configString, ",")
	}

	for _, config := range configs {
		log.Printf("Will load config %s", config)
	}

	var secretLocations []string
	if secretLocationsString != "" {
		secretLocations = strings.Split(secretLocationsString, ",")
	}

	// Load configs
	if err := load.Configs(configs, customConfig, secretLocations); err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	return nil
}
