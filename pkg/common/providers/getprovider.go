package providers

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"reflect"

	"github.com/markbates/pkger"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
)

const (
	// OCM provider.
	OCM = "ocm"

	// Mock provider.
	Mock = "mock"
)

// At the start of time, we'll populate the provider creation functions.
var providerCreationFunctions = map[string](func(cfg *config.Config) (spi.Provider, error)){}

func init() {
	cfg := config.Instance

	if cfg.Plugins != "" {
		err := filepath.Walk(cfg.Plugins, func(path string, info os.FileInfo, err error) error {
			if info != nil && !info.IsDir() {
				err := loadPlugin(path)

				if err != nil {
					log.Printf("encountered err: %v", err)
					return err
				}
			}

			return nil
		})

		if err != nil {
			panic(fmt.Sprintf("error loading user supplied plugins: %v", err))
		}
	}

	// Walk the built in plugins and attempt to load them as well.
	err := pkger.Walk("/assets/plugins", func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			file, err := pkger.Open(path)

			if err != nil {
				return fmt.Errorf("error opening built in plugin: %v", err)
			}

			// Create temporary file
			tmpFile, err := ioutil.TempFile("", "")

			if err != nil {
				return fmt.Errorf("unable to create temporary file: %v", err)
			}

			_, err = io.Copy(tmpFile, file)

			if err != nil {
				return fmt.Errorf("error copying built in plugin to temp file: %v", err)
			}

			defer func() {
				err := tmpFile.Close()

				if err != nil {
					log.Printf("error closing temporary file: %v", err)
				}

				err = os.Remove(tmpFile.Name())

				if err != nil {
					log.Printf("error removing temporary file: %v", err)
				}
			}()

			err = loadPlugin(tmpFile.Name())

			if err != nil {
				log.Printf("encountered err: %v", err)
				return err
			}
		}

		return nil
	})

	if err != nil {
		panic(fmt.Sprintf("error loading built in plugins: %v", err))
	}
}

func loadPlugin(path string) error {
	providerPlugin, err := plugin.Open(path)

	if err != nil {
		return fmt.Errorf("error loading plugin: %v", err)
	}

	nameFuncSym, err := providerPlugin.Lookup("ProviderName")

	if err != nil {
		return fmt.Errorf("error finding Name() function in plugin: %v", err)
	}

	nameFunc, ok := nameFuncSym.(func() string)

	if !ok {
		return fmt.Errorf("Name() is not of expected function signature (func() string)")
	}

	providerName := nameFunc()

	newFuncSym, err := providerPlugin.Lookup("New")

	if err != nil {
		return fmt.Errorf("error finding New() function in plugin: %v", err)
	}

	newFunc, ok := newFuncSym.(func(cfg *config.Config) (spi.Provider, error))

	if !ok {
		return fmt.Errorf("New() is not of expected function signature (func(cfg *config.Config) (spi.Provider, error). Actual: %s", reflect.TypeOf(newFuncSym))
	}

	if _, ok := providerCreationFunctions[providerName]; ok {
		log.Printf("Duplicate plugin with provider name: %s. Skipping.", providerName)
		return nil

	}

	providerCreationFunctions[providerName] = newFunc

	log.Printf("Loading plugin for %s.", providerName)

	return nil
}

// ClusterProvider returns the provisioner configured by the config object.
func ClusterProvider() (spi.Provider, error) {
	cfg := config.Instance
	providerName := cfg.Provider

	if providerNew, ok := providerCreationFunctions[providerName]; ok {
		return providerNew(cfg)
	}

	return nil, fmt.Errorf("unable to find provider %s", providerName)
}
