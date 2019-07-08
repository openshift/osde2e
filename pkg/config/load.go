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
		if env, ok := f.Tag.Lookup("env"); ok {
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
					if num, err := strconv.ParseInt(envVal, 10, 0); err == nil {
						field.SetInt(num)
					}
				}
			}
		}
	}
}
