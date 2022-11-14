package load

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/osde2e/configs"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
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
		for _, secret := range secrets {
			loadSecretFileIntoKey(secret.Key, secret.FileLocation, secretLocations)
		}

		for _, folder := range secretLocations {
			passthruSecrets := viper.GetStringMapString(config.NonOSDe2eSecrets)
			// Omit the osde2e secrets from going to the pass through secrets.
			if strings.Contains(folder, "osde2e-credentials") || strings.Contains(folder, "osde2e-common") {
				// If this is an addon test we will want to pass the ocm-token through.
				if viper.Get(config.Addons.IDs) != nil {
					_, exist := passthruSecrets["ocm-token-refresh"]
					if !exist {
						passthruSecrets["ocm-token-refresh"] = viper.GetString("ocm.token")
						passthruSecrets["ENV"] = viper.GetString("ocm.env")
						viper.Set(config.NonOSDe2eSecrets, passthruSecrets)
					}
				}
				continue
			}

			err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				if err != nil {
					return fmt.Errorf("Error walking folder %s: %s", folder, err.Error())
				}
				data, err := ioutil.ReadFile(path)
				if err != nil {
					return fmt.Errorf("error loading passthru-secret file %s", path)
				}
				passthruSecrets[info.Name()] = strings.TrimSpace(string(data))
				return nil
			})
			if err != nil {
				log.Printf("Error loading secret: %s", err.Error())
			}
			viper.Set(config.NonOSDe2eSecrets, passthruSecrets)
		}

	}

	// 4. Config post-processing.
	config.PostProcess()

	return nil
}

// loadYAMLFromConfigs accepts a config name and attempts to unmarshal the config from the /configs directory.
func loadYAMLFromConfigs(name string) error {
	var (
		file fs.File
		err  error
	)

	if file, err = configs.FS.Open(name + ".yaml"); err != nil {
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
	// We should rewrite all of this logic. This introduces a bug with the current expected behavior that overwrites values but does this multiple times.
	for _, secretLocation := range secretLocations {
		fullFilename := filepath.Join(secretLocation, filename)
		// This is a bandage fix until we can rewrite the logic to load secrets.
		if (strings.Contains(fullFilename, "osde2e-credentials") || strings.Contains(fullFilename, "osde2e-common")) && (key == "ocm.aws.accesKey" || key == "ocm.aws.secretKey") {
			if viper.GetBool("ocm.ccs") {
				continue
			}
		}

		stat, err := os.Stat(fullFilename)
		if err == nil && !stat.IsDir() {
			data, err := ioutil.ReadFile(fullFilename)
			if err != nil {
				return fmt.Errorf("error loading secret file %s from location %s", filename, secretLocation)
			}
			log.Printf("Found secret for key %s.", key)
			cleanData := strings.TrimSpace(string(data))
			if cleanData != "" {
				// If the data contains a certificate, we'll need to pass the file path to the secret.
				if strings.Contains(cleanData, "-----BEGIN CERTIFICATE-----") {
					cleanData = fullFilename
				}
				viper.Set(key, cleanData)
			}
			return nil
		}
	}

	return nil
}
