package load

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/markbates/pkger"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/spf13/viper"
)

const (
	// EnvVarTag is the Go struct tag containing the environment variable that sets the option.
	EnvVarTag = "env"

	// SectionTag is the Go struct tag containing the documentation section of the option.
	SectionTag = "sect"

	// DefaultTag is the Go struct tag containing the default value of the option.
	DefaultTag = "default"
)

// This is a set of pre-canned configs that will always be loaded at startup.
var defaultConfigs = []string{
	"log-metrics",
	"before-suite-metrics",
	"aws-log-metrics",
	"aws-before-suite-metrics",
	"gcp-before-suite-metrics",
	"gcp-log-metrics",
}

// Configs will populate viper with specified configs.
func Configs(configs []string, customConfig string, secretLocations []string) error {
	// This used to be complicated, but now we just lean on Viper for everything.
	// 1. Load default configs. These are configs that will always be enabled for every run.
	for _, config := range defaultConfigs {
		if err := loadYAMLFromConfigs(config); err != nil {
			return fmt.Errorf("error loading config from YAML: %v", err)
		}
	}

	// 2. Load pre-canned YAML configs.
	for _, config := range configs {
		if err := loadYAMLFromConfigs(config); err != nil {
			return fmt.Errorf("error loading config from YAML: %v", err)
		}
	}

	// 3. Custom YAML configs
	if customConfig != "" {
		log.Printf("Custom YAML config provided, loading from %s", customConfig)
		if err := loadYAMLFromFile(customConfig); err != nil {
			return fmt.Errorf("error loading custom config from YAML: %v", err)
		}
	}

	// 4. Secrets. These will override all previous entries.
	if len(secretLocations) > 0 {
		secrets := config.GetAllSecrets()
		for key, secretFilename := range secrets {
			loadSecretFileIntoKey(key, secretFilename, secretLocations)
		}
	}

	// 4. Config post-processing.
	config.PostProcess()

	return nil
}

// loadYAMLFromConfigs accepts a config name and attempts to unmarshal the config from the /configs directory.
func loadYAMLFromConfigs(name string) error {
	var file http.File
	var err error

	if file, err = pkger.Open(filepath.Join("/configs", name+".yaml")); err != nil {
		return fmt.Errorf("error trying to open config %s: %v", name, err)
	}

	defer file.Close()

	if err = viper.MergeConfig(file); err != nil {
		return err
	}

	return nil
}

// loadYAMLFromFile accepts file info and attempts to unmarshal the file into the // config.
func loadYAMLFromFile(name string) error {
	var err error
	var dir, path string

	if dir, err = os.Getwd(); err != nil {
		log.Fatalf("Unable to get CWD: %s", err.Error())
	}
	// TODO: This needs to change once we stop branching out execution the way we do it currently
	// It's fragile
	if path, err = filepath.Abs(filepath.Join(dir, name)); err != nil {
		return err
	}

	path = filepath.Clean(path)

	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	if err = viper.MergeConfig(fh); err != nil {
		return err
	}

	return nil
}

// loadSecretFileIntoKey will attempt to load the contents of a secret file into the given key.
// If the secret file doesn't exist, we'll skip this.
func loadSecretFileIntoKey(key string, filename string, secretLocations []string) error {
	for _, secretLocation := range secretLocations {
		fullFilename := filepath.Join(secretLocation, filename)
		stat, err := os.Stat(fullFilename)
		if err == nil && !stat.IsDir() {
			data, err := ioutil.ReadFile(fullFilename)
			if err != nil {
				return fmt.Errorf("error loading secret file %s from location %s", filename, secretLocation)
			}
			log.Printf("Found secret for key %s.", key)
			viper.Set(key, strings.TrimSpace(string(data)))
			return nil
		}
	}

	return nil
}
