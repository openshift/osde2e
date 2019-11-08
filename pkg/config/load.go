package config

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
)

func init() {
	if err := Cfg.LoadFromEnv(); err != nil {
		log.Fatal("error loading config from environment: ", err)
	}
}

// LoadFromEnv sets values from environment variables specified in `env` tags.
func (c *Config) LoadFromEnv() error {
	v := reflect.ValueOf(c).Elem()
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		if defaultValue, ok := f.Tag.Lookup(DefaultTag); ok {
			field := v.Field(i)
			switch f.Type.Kind() {
			case reflect.String:
				field.SetString(defaultValue)
			case reflect.Bool:
				if newBool, err := strconv.ParseBool(defaultValue); err == nil {
					field.SetBool(newBool)
				} else {
					return fmt.Errorf("error parsing bool value for field %s: %v", f.Name, err)
				}
			case reflect.Slice:
				field.SetBytes([]byte(defaultValue))
			case reflect.Int:
				fallthrough
			case reflect.Int64:
				if num, err := strconv.ParseInt(defaultValue, 10, 0); err == nil {
					field.SetInt(num)
				} else {
					return fmt.Errorf("error parsing int value for field %s: %v", f.Name, err)
				}
			}
		}
		if env, ok := f.Tag.Lookup(EnvVarTag); ok {
			if envVal := os.Getenv(env); len(envVal) > 0 {
				field := v.Field(i)
				switch f.Type.Kind() {
				case reflect.String:
					field.SetString(envVal)
				case reflect.Bool:
					if newBool, err := strconv.ParseBool(envVal); err == nil {
						field.SetBool(newBool)
					} else {
						return fmt.Errorf("error parsing bool value for field %s: %v", f.Name, err)
					}
				case reflect.Slice:
					field.SetBytes([]byte(envVal))
				case reflect.Int:
					fallthrough
				case reflect.Int64:
					if num, err := strconv.ParseInt(envVal, 10, 0); err == nil {
						field.SetInt(num)
					} else {
						return fmt.Errorf("error parsing int value for field %s: %v", f.Name, err)
					}
				}
			}
		}
	}

	return nil
}
