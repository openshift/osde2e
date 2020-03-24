package load

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/markbates/pkger"
	"github.com/openshift/osde2e/pkg/common/util"
	"gopkg.in/yaml.v2"
)

const (
	// EnvVarTag is the Go struct tag containing the environment variable that sets the option.
	EnvVarTag = "env"

	// SectionTag is the Go struct tag containing the documentation section of the option.
	SectionTag = "sect"

	// DefaultTag is the Go struct tag containing the default value of the option.
	DefaultTag = "default"

	// AppendYamlArrayTag is the Go struct tag that instructs the config loader to append arrays instead of overwriting them when loading YAML.
	AppendYamlArrayTag = "appendYamlArray"
)

var rndStringRegex = regexp.MustCompile("__RND_(\\d+)__")

func init() {
	rand.Seed(time.Now().Unix())
}

// IntoObject populates an object based on the tags specified in the object.
func IntoObject(object interface{}, configs []string, customConfig string) error {
	if objectType := reflect.TypeOf(object); objectType.Kind() != reflect.Ptr {
		return fmt.Errorf("the supplied object must be a pointer")
	}

	// Populate the defaults first, then read the YAML, then override with the environment
	if err := loadDefaults(object); err != nil {
		return fmt.Errorf("error loading config defaults: %v", err)
	}

	for _, config := range configs {
		if err := loadYAMLFromConfigs(object, config); err != nil {
			return fmt.Errorf("error loading config from YAML: %v", err)
		}
	}

	if customConfig != "" {
		log.Printf("Custom YAML config provided, loading from %s", customConfig)
		if err := loadYAMLFromFile(object, customConfig); err != nil {
			return fmt.Errorf("error loading custom config from YAML: %v", err)
		}
	}

	if err := loadFromEnv(object); err != nil {
		return fmt.Errorf("error loading config from environment: %v", err)
	}

	return nil
}

// load values into the given field
func load(v reflect.Value, source string) error {
	var setValue string
	var ok bool
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)

		if f.Type.Kind() == reflect.Struct {
			// Specific to supporting AddOns via ENV
			load(v.FieldByIndex(f.Index), source)
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

// loadDefaults takes default values from the annotations in the types
// file and assigns them to the appropriate config option.
// It also works on handling special cases for default loading.
func loadDefaults(object interface{}) error {
	v := reflect.ValueOf(object).Elem()
	load(v, "default")
	return nil
}

// loadYAMLFromConfigs accepts a config name and attempts to unmarshal the config from the /configs directory.
func loadYAMLFromConfigs(object interface{}, name string) error {
	var file http.File
	var data []byte
	var err error

	if file, err = pkger.Open(filepath.Join("/configs", name+".yaml")); err != nil {
		return err
	}

	if data, err = ioutil.ReadAll(file); err != nil {
		return err
	}

	if err = loadYAMLFromData(object, data); err != nil {
		return err
	}

	return nil
}

// loadYAMLFromFile accepts file info and attempts to unmarshal the file into the // config.
func loadYAMLFromFile(object interface{}, name string) error {
	var data []byte
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

	if data, err = ioutil.ReadFile(path); err != nil {
		return err
	}

	if err = loadYAMLFromData(object, data); err != nil {
		return err
	}

	return nil
}

// loadYAMLFromData will take a byte array and load YAML into the given object
func loadYAMLFromData(object interface{}, data []byte) (err error) {
	objectType := reflect.Indirect(reflect.ValueOf(object)).Type()
	tempObject := reflect.New(objectType).Interface()

	if err = yaml.Unmarshal(data, tempObject); err != nil {
		return err
	}

	if err = copyFromTempObject(tempObject, object); err != nil {
		return fmt.Errorf("error copying from temp object to actual object: %v", err)
	}

	return nil
}

// copyFromTempObject copies from a temporary object into the actual object.
// The temp object is expected to not have any defaults loaded into it, as any "zero" valued
// fields in the temp object will not be assigned to the object. This will prevent values from
// being inadvertently overwritten in the actual object.
func copyFromTempObject(tempObject, object interface{}) error {
	// We need a variety of reflections to start. These are for detecting the object type underneath any
	// pointers and to get and set the eventual values of the fields.
	tempObjectValue := reflect.Indirect(reflect.ValueOf(tempObject))
	objectValue := reflect.Indirect(reflect.ValueOf(object))
	tempObjectType := tempObjectValue.Type()
	objectType := objectValue.Type()

	if tempObjectType != objectType && tempObjectType.Kind() != reflect.Struct {
		return fmt.Errorf("temp object and destination object types must be the same (src: %v, dst: %v) and must be structs", tempObjectType, objectType)
	}

	for i := 0; i < tempObjectType.NumField(); i++ {
		// Getting the StructField value from (temp)ObjectType here will allow us to retrieve tag values.
		tempObjectStructField := tempObjectType.Field(i)
		objectStructField := objectType.Field(i)
		// Getting the Values from the (temp)ObjectValues will allow us to get and set the field values.
		tempObjectField := tempObjectValue.Field(i)
		objectField := objectValue.Field(i)

		if !objectField.CanSet() {
			continue
		}

		if tempObjectStructField.Name != objectStructField.Name {
			return fmt.Errorf("source and destination fields do not match during copy (src: %v, dst: %v)", tempObjectField, objectField)
		}

		// Here we're looking for a special tag that tells us not to overwrite an array value with the YAML value in the
		// temp object, but to append it onto the value currently on the actual object. This will allow us to do things like
		// have multiple test suites specified simultaneously that will not overwrite one another.
		if append, _ := strconv.ParseBool(tempObjectStructField.Tag.Get(AppendYamlArrayTag)); append {
			objectField.Set(reflect.AppendSlice(objectField, tempObjectField))
		} else if tempObjectStructField.Type.Kind() == reflect.Struct {
			// Recursively call this function on member structs.
			if err := copyFromTempObject(tempObjectField.Addr().Interface(), objectField.Addr().Interface()); err != nil {
				return fmt.Errorf("error loading struct %s: %v", tempObjectField.Type().Name(), err)
			}
		} else {
			// If the temp object field is equal to the zero value for the corresponding type, we will not set the corresponding
			// field in the actual object. Since we're able to compose multiple YAML files together, small YAML files with
			// a few directives will contain a lot of zero values (empty strings, integers == 0, etc.). If we allow zero
			// values to clobber the fields in the actual object, each small YAML config file that we use will potentially "unset"
			// the values intentionally set by other YAML files.
			//
			// In other words, if the user hasn't set a field in a YAML config, don't try to set it in the actual object.
			if !reflect.DeepEqual(tempObjectField.Interface(), reflect.Zero(tempObjectField.Type()).Interface()) {
				objectField.Set(tempObjectField)
			}
		}
	}

	return nil
}

// loadFromEnv sets values from environment variables specified in `env` tags.
func loadFromEnv(object interface{}) error {
	v := reflect.ValueOf(object).Elem()
	load(v, "env")

	return nil
}

func processValueFromString(f reflect.StructField, field reflect.Value, value string) error {
	switch f.Type.Kind() {
	case reflect.String:
		// Add special processing for the __TMP_DIR__ string so that directory creation is handled
		// internally to config loading.
		if value == "__TMP_DIR__" {
			if dir, err := ioutil.TempDir("", "osde2e"); err == nil {
				log.Printf("Generated temporary directory %s for field %s", dir, f.Name)
				field.SetString(dir)
			} else {
				return fmt.Errorf("error generating temporary directory for field %s: %v", f.Name, err)
			}
		} else if rndStringRegex.MatchString(value) {
			if rndStringLen, err := strconv.Atoi(rndStringRegex.FindStringSubmatch(value)[1]); err == nil {
				rndString := util.RandomStr(rndStringLen)
				log.Printf("Generated random string %s for field %s", rndString, f.Name)
				field.SetString(rndString)
			} else {
				return fmt.Errorf("error generating random string for field %s: %v", f.Name, err)
			}
		} else {
			field.SetString(value)
		}
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
