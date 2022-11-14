//nolint
package concurrentviper

import (
	io "io"
	os "os"
	strings "strings"
	sync "sync"
	time "time"

	fsnotify "github.com/fsnotify/fsnotify"
	mapstructure "github.com/mitchellh/mapstructure"
	afero "github.com/spf13/afero"
	pflag "github.com/spf13/pflag"
	viper "github.com/spf13/viper"
)

var l sync.Mutex

// DecodeHook returns a DecoderConfigOption which overrides the default
// DecoderConfig.DecodeHook value, the default is:
//
//  mapstructure.ComposeDecodeHookFunc(
//		mapstructure.StringToTimeDurationHookFunc(),
//		mapstructure.StringToSliceHookFunc(","),
//	)
// This function is safe for concurrent use.
func DecodeHook(hook mapstructure.DecodeHookFunc) viper.DecoderConfigOption {
	l.Lock()
	defer l.Unlock()
	return viper.DecodeHook(hook)
}

// New returns an initialized Viper instance.
// This function is safe for concurrent use.
func New() *viper.Viper { l.Lock(); defer l.Unlock(); return viper.New() }

// KeyDelimiter sets the delimiter used for determining key parts.
// By default it's value is ".".
// This function is safe for concurrent use.
func KeyDelimiter(d string) viper.Option { l.Lock(); defer l.Unlock(); return viper.KeyDelimiter(d) }

// EnvKeyReplacer sets a replacer used for mapping environment variables to internal keys.
// This function is safe for concurrent use.
func EnvKeyReplacer(r viper.StringReplacer) viper.Option {
	l.Lock()
	defer l.Unlock()
	return viper.EnvKeyReplacer(r)
}

// NewWithOptions creates a new Viper instance.
// This function is safe for concurrent use.
func NewWithOptions(opts ...viper.Option) *viper.Viper {
	l.Lock()
	defer l.Unlock()
	return viper.NewWithOptions(opts...)
}

// Reset is intended for testing, will reset all to default settings.
// In the public interface for the viper package so applications
// can use it in their testing as well.
// This function is safe for concurrent use.
func Reset() { l.Lock(); defer l.Unlock(); viper.Reset() }

// This function is safe for concurrent use.
func OnConfigChange(run func(in fsnotify.Event)) {
	l.Lock()
	defer l.Unlock()
	viper.OnConfigChange(run)
}

// This function is safe for concurrent use.
func WatchConfig() { l.Lock(); defer l.Unlock(); viper.WatchConfig() }

// SetConfigFile explicitly defines the path, name and extension of the config file.
// Viper will use this and not check any of the config paths.
// This function is safe for concurrent use.
func SetConfigFile(in string) { l.Lock(); defer l.Unlock(); viper.SetConfigFile(in) }

// SetEnvPrefix defines a prefix that ENVIRONMENT variables will use.
// E.g. if your prefix is "spf", the env registry will look for env
// variables that start with "SPF_".
// This function is safe for concurrent use.
func SetEnvPrefix(in string) { l.Lock(); defer l.Unlock(); viper.SetEnvPrefix(in) }

// AllowEmptyEnv tells Viper to consider set,
// but empty environment variables as valid values instead of falling back.
// For backward compatibility reasons this is false by default.
// This function is safe for concurrent use.
func AllowEmptyEnv(allowEmptyEnv bool) {
	l.Lock()
	defer l.Unlock()
	viper.AllowEmptyEnv(allowEmptyEnv)
}

// ConfigFileUsed returns the file used to populate the config registry.
// This function is safe for concurrent use.
func ConfigFileUsed() string { l.Lock(); defer l.Unlock(); return viper.ConfigFileUsed() }

// AddConfigPath adds a path for Viper to search for the config file in.
// Can be called multiple times to define multiple search paths.
// This function is safe for concurrent use.
func AddConfigPath(in string) { l.Lock(); defer l.Unlock(); viper.AddConfigPath(in) }

// AddRemoteProvider adds a remote configuration source.
// Remote Providers are searched in the order they are added.
// provider is a string value: "etcd", "consul" or "firestore" are currently supported.
// endpoint is the url.  etcd requires http://ip:port  consul requires ip:port
// path is the path in the k/v store to retrieve configuration
// To retrieve a config file called myapp.json from /configs/myapp.json
// you should set path to /configs and set config name (SetConfigName()) to
// "myapp"
// This function is safe for concurrent use.
func AddRemoteProvider(provider, endpoint, path string) error {
	l.Lock()
	defer l.Unlock()
	return viper.AddRemoteProvider(provider, endpoint, path)
}

// AddSecureRemoteProvider adds a remote configuration source.
// Secure Remote Providers are searched in the order they are added.
// provider is a string value: "etcd", "consul" or "firestore" are currently supported.
// endpoint is the url.  etcd requires http://ip:port  consul requires ip:port
// secretkeyring is the filepath to your openpgp secret keyring.  e.g. /etc/secrets/myring.gpg
// path is the path in the k/v store to retrieve configuration
// To retrieve a config file called myapp.json from /configs/myapp.json
// you should set path to /configs and set config name (SetConfigName()) to
// "myapp"
// Secure Remote Providers are implemented with github.com/bketelsen/crypt
// This function is safe for concurrent use.
func AddSecureRemoteProvider(provider, endpoint, path, secretkeyring string) error {
	l.Lock()
	defer l.Unlock()
	return viper.AddSecureRemoteProvider(provider, endpoint, path, secretkeyring)
}

// SetTypeByDefaultValue enables or disables the inference of a key value's
// type when the Get function is used based upon a key's default value as
// opposed to the value returned based on the normal fetch logic.
//
// For example, if a key has a default value of []string{} and the same key
// is set via an environment variable to "a b c", a call to the Get function
// would return a string slice for the key if the key's type is inferred by
// the default value and the Get function would return:
//
//   []string {"a", "b", "c"}
//
// Otherwise the Get function would return:
//
//   "a b c"
// This function is safe for concurrent use.
func SetTypeByDefaultValue(enable bool) {
	l.Lock()
	defer l.Unlock()
	viper.SetTypeByDefaultValue(enable)
}

// GetViper gets the global Viper instance.
// This function is safe for concurrent use.
func GetViper() *viper.Viper { l.Lock(); defer l.Unlock(); return viper.GetViper() }

// Get can retrieve any value given the key to use.
// Get is case-insensitive for a key.
// Get has the behavior of returning the value associated with the first
// place from where it is set. Viper will check in the following order:
// override, flag, env, config file, key/value store, default
//
// Get returns an interface. For a specific value use one of the Get____ methods.
// This function is safe for concurrent use.
func Get(key string) interface{} { l.Lock(); defer l.Unlock(); return viper.Get(key) }

// Sub returns new Viper instance representing a sub tree of this instance.
// Sub is case-insensitive for a key.
// This function is safe for concurrent use.
func Sub(key string) *viper.Viper { l.Lock(); defer l.Unlock(); return viper.Sub(key) }

// GetString returns the value associated with the key as a string.
// This function is safe for concurrent use.
func GetString(key string) string { l.Lock(); defer l.Unlock(); return viper.GetString(key) }

// GetBool returns the value associated with the key as a boolean.
// This function is safe for concurrent use.
func GetBool(key string) bool { l.Lock(); defer l.Unlock(); return viper.GetBool(key) }

// GetInt returns the value associated with the key as an integer.
// This function is safe for concurrent use.
func GetInt(key string) int { l.Lock(); defer l.Unlock(); return viper.GetInt(key) }

// GetInt32 returns the value associated with the key as an integer.
// This function is safe for concurrent use.
func GetInt32(key string) int32 { l.Lock(); defer l.Unlock(); return viper.GetInt32(key) }

// GetInt64 returns the value associated with the key as an integer.
// This function is safe for concurrent use.
func GetInt64(key string) int64 { l.Lock(); defer l.Unlock(); return viper.GetInt64(key) }

// GetUint returns the value associated with the key as an unsigned integer.
// This function is safe for concurrent use.
func GetUint(key string) uint { l.Lock(); defer l.Unlock(); return viper.GetUint(key) }

// GetUint32 returns the value associated with the key as an unsigned integer.
// This function is safe for concurrent use.
func GetUint32(key string) uint32 { l.Lock(); defer l.Unlock(); return viper.GetUint32(key) }

// GetUint64 returns the value associated with the key as an unsigned integer.
// This function is safe for concurrent use.
func GetUint64(key string) uint64 { l.Lock(); defer l.Unlock(); return viper.GetUint64(key) }

// GetFloat64 returns the value associated with the key as a float64.
// This function is safe for concurrent use.
func GetFloat64(key string) float64 { l.Lock(); defer l.Unlock(); return viper.GetFloat64(key) }

// GetTime returns the value associated with the key as time.
// This function is safe for concurrent use.
func GetTime(key string) time.Time { l.Lock(); defer l.Unlock(); return viper.GetTime(key) }

// GetDuration returns the value associated with the key as a duration.
// This function is safe for concurrent use.
func GetDuration(key string) time.Duration { l.Lock(); defer l.Unlock(); return viper.GetDuration(key) }

// GetIntSlice returns the value associated with the key as a slice of int values.
// This function is safe for concurrent use.
func GetIntSlice(key string) []int { l.Lock(); defer l.Unlock(); return viper.GetIntSlice(key) }

// GetStringSlice returns the value associated with the key as a slice of strings.
// This function is safe for concurrent use.
func GetStringSlice(key string) []string {
	l.Lock()
	defer l.Unlock()
	return viper.GetStringSlice(key)
}

// GetStringMap returns the value associated with the key as a map of interfaces.
// This function is safe for concurrent use.
func GetStringMap(key string) map[string]interface{} {
	l.Lock()
	defer l.Unlock()
	return viper.GetStringMap(key)
}

// GetStringMapString returns the value associated with the key as a map of strings.
// This function is safe for concurrent use.
func GetStringMapString(key string) map[string]string {
	l.Lock()
	defer l.Unlock()
	return viper.GetStringMapString(key)
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
// This function is safe for concurrent use.
func GetStringMapStringSlice(key string) map[string][]string {
	l.Lock()
	defer l.Unlock()
	return viper.GetStringMapStringSlice(key)
}

// GetSizeInBytes returns the size of the value associated with the given key
// in bytes.
// This function is safe for concurrent use.
func GetSizeInBytes(key string) uint { l.Lock(); defer l.Unlock(); return viper.GetSizeInBytes(key) }

// UnmarshalKey takes a single key and unmarshals it into a Struct.
// This function is safe for concurrent use.
func UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	l.Lock()
	defer l.Unlock()
	return viper.UnmarshalKey(key, rawVal, opts...)
}

// Unmarshal unmarshals the config into a Struct. Make sure that the tags
// on the fields of the structure are properly set.
// This function is safe for concurrent use.
func Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	l.Lock()
	defer l.Unlock()
	return viper.Unmarshal(rawVal, opts...)
}

// UnmarshalExact unmarshals the config into a Struct, erroring if a field is nonexistent
// in the destination struct.
// This function is safe for concurrent use.
func UnmarshalExact(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	l.Lock()
	defer l.Unlock()
	return viper.UnmarshalExact(rawVal, opts...)
}

// BindPFlags binds a full flag set to the configuration, using each flag's long
// name as the config key.
// This function is safe for concurrent use.
func BindPFlags(flags *pflag.FlagSet) error {
	l.Lock()
	defer l.Unlock()
	return viper.BindPFlags(flags)
}

// BindPFlag binds a specific key to a pflag (as used by cobra).
// Example (where serverCmd is a Cobra instance):
//
//	 serverCmd.Flags().Int("port", 1138, "Port to run Application server on")
//	 Viper.BindPFlag("port", serverCmd.Flags().Lookup("port"))
//
// This function is safe for concurrent use.
func BindPFlag(key string, flag *pflag.Flag) error {
	l.Lock()
	defer l.Unlock()
	return viper.BindPFlag(key, flag)
}

// BindFlagValues binds a full FlagValue set to the configuration, using each flag's long
// name as the config key.
// This function is safe for concurrent use.
func BindFlagValues(flags viper.FlagValueSet) error {
	l.Lock()
	defer l.Unlock()
	return viper.BindFlagValues(flags)
}

// BindFlagValue binds a specific key to a FlagValue.
// This function is safe for concurrent use.
func BindFlagValue(key string, flag viper.FlagValue) error {
	l.Lock()
	defer l.Unlock()
	return viper.BindFlagValue(key, flag)
}

// BindEnv binds a Viper key to a ENV variable.
// ENV variables are case sensitive.
// If only a key is provided, it will use the env key matching the key, uppercased.
// EnvPrefix will be used when set when env name is not provided.
// This function is safe for concurrent use.
func BindEnv(input ...string) error { l.Lock(); defer l.Unlock(); return viper.BindEnv(input...) }

// IsSet checks to see if the key has been set in any of the data locations.
// IsSet is case-insensitive for a key.
// This function is safe for concurrent use.
func IsSet(key string) bool { l.Lock(); defer l.Unlock(); return viper.IsSet(key) }

// AutomaticEnv has Viper check ENV variables for all.
// keys set in config, default & flags
// This function is safe for concurrent use.
func AutomaticEnv() { l.Lock(); defer l.Unlock(); viper.AutomaticEnv() }

// SetEnvKeyReplacer sets the strings.Replacer on the viper object
// Useful for mapping an environmental variable to a key that does
// not match it.
// This function is safe for concurrent use.
func SetEnvKeyReplacer(r *strings.Replacer) { l.Lock(); defer l.Unlock(); viper.SetEnvKeyReplacer(r) }

// RegisterAlias creates an alias that provides another accessor for the same key.
// This enables one to change a name without breaking the application.
// This function is safe for concurrent use.
func RegisterAlias(alias string, key string) {
	l.Lock()
	defer l.Unlock()
	viper.RegisterAlias(alias, key)
}

// InConfig checks to see if the given key (or an alias) is in the config file.
// This function is safe for concurrent use.
func InConfig(key string) bool { l.Lock(); defer l.Unlock(); return viper.InConfig(key) }

// SetDefault sets the default value for this key.
// SetDefault is case-insensitive for a key.
// Default only used when no value is provided by the user via flag, config or ENV.
// This function is safe for concurrent use.
func SetDefault(key string, value interface{}) {
	l.Lock()
	defer l.Unlock()
	viper.SetDefault(key, value)
}

// Set sets the value for the key in the override register.
// Set is case-insensitive for a key.
// Will be used instead of values obtained via
// flags, config file, ENV, default, or key/value store.
// This function is safe for concurrent use.
func Set(key string, value interface{}) { l.Lock(); defer l.Unlock(); viper.Set(key, value) }

// ReadInConfig will discover and load the configuration file from disk
// and key/value stores, searching in one of the defined paths.
// This function is safe for concurrent use.
func ReadInConfig() error { l.Lock(); defer l.Unlock(); return viper.ReadInConfig() }

// MergeInConfig merges a new configuration with an existing config.
// This function is safe for concurrent use.
func MergeInConfig() error { l.Lock(); defer l.Unlock(); return viper.MergeInConfig() }

// ReadConfig will read a configuration file, setting existing keys to nil if the
// key does not exist in the file.
// This function is safe for concurrent use.
func ReadConfig(in io.Reader) error { l.Lock(); defer l.Unlock(); return viper.ReadConfig(in) }

// MergeConfig merges a new configuration with an existing config.
// This function is safe for concurrent use.
func MergeConfig(in io.Reader) error { l.Lock(); defer l.Unlock(); return viper.MergeConfig(in) }

// MergeConfigMap merges the configuration from the map given with an existing config.
// Note that the map given may be modified.
// This function is safe for concurrent use.
func MergeConfigMap(cfg map[string]interface{}) error {
	l.Lock()
	defer l.Unlock()
	return viper.MergeConfigMap(cfg)
}

// WriteConfig writes the current configuration to a file.
// This function is safe for concurrent use.
func WriteConfig() error { l.Lock(); defer l.Unlock(); return viper.WriteConfig() }

// SafeWriteConfig writes current configuration to file only if the file does not exist.
// This function is safe for concurrent use.
func SafeWriteConfig() error { l.Lock(); defer l.Unlock(); return viper.SafeWriteConfig() }

// WriteConfigAs writes current configuration to a given filename.
// This function is safe for concurrent use.
func WriteConfigAs(filename string) error {
	l.Lock()
	defer l.Unlock()
	return viper.WriteConfigAs(filename)
}

// SafeWriteConfigAs writes current configuration to a given filename if it does not exist.
// This function is safe for concurrent use.
func SafeWriteConfigAs(filename string) error {
	l.Lock()
	defer l.Unlock()
	return viper.SafeWriteConfigAs(filename)
}

// ReadRemoteConfig attempts to get configuration from a remote source
// and read it in the remote configuration registry.
// This function is safe for concurrent use.
func ReadRemoteConfig() error { l.Lock(); defer l.Unlock(); return viper.ReadRemoteConfig() }

// This function is safe for concurrent use.
func WatchRemoteConfig() error { l.Lock(); defer l.Unlock(); return viper.WatchRemoteConfig() }

// AllKeys returns all keys holding a value, regardless of where they are set.
// Nested keys are returned with a v.keyDelim separator
// This function is safe for concurrent use.
func AllKeys() []string { l.Lock(); defer l.Unlock(); return viper.AllKeys() }

// AllSettings merges all settings and returns them as a map[string]interface{}.
// This function is safe for concurrent use.
func AllSettings() map[string]interface{} { l.Lock(); defer l.Unlock(); return viper.AllSettings() }

// SetFs sets the filesystem to use to read configuration.
// This function is safe for concurrent use.
func SetFs(fs afero.Fs) { l.Lock(); defer l.Unlock(); viper.SetFs(fs) }

// SetConfigName sets name for the config file.
// Does not include extension.
// This function is safe for concurrent use.
func SetConfigName(in string) { l.Lock(); defer l.Unlock(); viper.SetConfigName(in) }

// SetConfigType sets the type of the configuration returned by the
// remote source, e.g. "json".
// This function is safe for concurrent use.
func SetConfigType(in string) { l.Lock(); defer l.Unlock(); viper.SetConfigType(in) }

// SetConfigPermissions sets the permissions for the config file.
// This function is safe for concurrent use.
func SetConfigPermissions(perm os.FileMode) {
	l.Lock()
	defer l.Unlock()
	viper.SetConfigPermissions(perm)
}

// Debug prints all configuration registries for debugging
// purposes.
// This function is safe for concurrent use.
func Debug() { l.Lock(); defer l.Unlock(); viper.Debug() }
