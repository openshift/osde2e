package load

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/markbates/pkger"
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

// Configs will populate viper with specified configs.
func Configs(configs []string, customConfig string) error {
	// This used to be complicated, but now we just lean on Viper for everything.
	// 1. Load pre-canned YAML configs.
	for _, config := range configs {
		if err := loadYAMLFromConfigs(config); err != nil {
			return fmt.Errorf("error loading config from YAML: %v", err)
		}
	}

	// 2. Custom YAML configs
	if customConfig != "" {
		log.Printf("Custom YAML config provided, loading from %s", customConfig)
		if err := loadYAMLFromFile(customConfig); err != nil {
			return fmt.Errorf("error loading custom config from YAML: %v", err)
		}
	}

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

	var file http.File
	if file, err = pkger.Open(path); err != nil {
		return fmt.Errorf("error trying to open config %s: %v", name, err)
	}

	defer file.Close()

	if err = viper.MergeConfig(file); err != nil {
		return err
	}

	return nil
}
