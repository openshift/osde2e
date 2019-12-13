package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

func init() {
	log.Print("Defaulting to environment variables for config")
	if err := Cfg.LoadFromEnv(); err != nil {
		log.Fatal("error loading config values: ", err)
	}
}

// Load loads things
func (c *Config) Load(v reflect.Value, source string) error {
	var setValue string
	var ok bool
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)

		if f.Type.Kind() == reflect.Struct {
			// Specific to supporting AddOns via ENV
			c.Load(v.FieldByIndex(f.Index), source)
		} else {
			if source == "default" {
				if setValue, ok = f.Tag.Lookup(DefaultTag); !ok {
					continue
				}
			}
			if source == "env" {
				if env, ok := f.Tag.Lookup(EnvVarTag); ok {
					if setValue = os.Getenv(env); setValue == "" {
						continue
					}
				}
			}

			field := v.Field(i)
			if err := processValueFromString(f, field, setValue); err != nil {
				return err
			}
		}
	}
	return nil
}

// LoadDefaults takes default values from the annotations in the types
// file and assigns them to the appropriate config option
func (c *Config) LoadDefaults() error {
	v := reflect.ValueOf(c).Elem()
	c.Load(v, "default")
	return nil
}

// LoadFromYAML accepts file info and attempts to unmarshal the file into the
// config.
func (c *Config) LoadFromYAML(name string) error {
	var data []byte
	var err error
	var dir, path string

	if dir, err = os.Getwd(); err != nil {
		log.Fatalf("Unable to get CWD: %s", err.Error())
	}
	// TODO: This needs to change once we stop branching out execution the way we do it currently
	// It's fragile
	if path, err = filepath.Abs(filepath.Join(dir, "../../", name)); err != nil {
		return err
	}

	path = filepath.Clean(path)

	if data, err = ioutil.ReadFile(path); err != nil {
		return err
	}

	if err = yaml.Unmarshal(data, c); err != nil {
		return err
	}

	return nil
}

// LoadFromEnv sets values from environment variables specified in `env` tags.
func (c *Config) LoadFromEnv() error {
	if err := c.LoadDefaults(); err != nil {
		return err
	}

	v := reflect.ValueOf(c).Elem()
	c.Load(v, "env")

	return nil
}

func processValueFromString(f reflect.StructField, field reflect.Value, value string) error {
	switch f.Type.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		if newBool, err := strconv.ParseBool(value); err == nil {
			field.SetBool(newBool)
		} else {
			return fmt.Errorf("error parsing bool value for field %s: %v", f.Name, err)
		}
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		if value != "" {
			value := string(value)
			a := strings.Split(value, ",")
			for i := range a {
				field.Set(reflect.Append(field, reflect.ValueOf(a[i])))
			}
		}
		// We shouldn't be setting any slices with string vars
		// Specifically, Addons and Kubeconfig Contents
	case reflect.Int:
		fallthrough
	case reflect.Int64:
		if num, err := strconv.ParseInt(value, 10, 0); err == nil {
			field.SetInt(num)
		} else {
			return fmt.Errorf("error parsing int value for field %s: %v", f.Name, err)
		}
	}
	return nil
}
