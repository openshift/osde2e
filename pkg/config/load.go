package config

import (
	"os"
	"reflect"
	"strconv"
)

func init() {
	Cfg.LoadFromEnv()
}

// LoadFromEnv sets values from environment variables specified in `env` tags.
func (c *Config) LoadFromEnv() {
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
				}
			case reflect.Slice:
				field.SetBytes([]byte(defaultValue))
			case reflect.Int:
				fallthrough
			case reflect.Int64:
				if num, err := strconv.ParseInt(defaultValue, 10, 0); err == nil {
					field.SetInt(num)
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
					field.SetBool(true)
				case reflect.Slice:
					field.SetBytes([]byte(envVal))
				case reflect.Int:
					fallthrough
				case reflect.Int64:
					if num, err := strconv.ParseInt(envVal, 10, 0); err == nil {
						field.SetInt(num)
					}
				}
			}
		}
	}
}
