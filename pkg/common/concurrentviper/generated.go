package concurrentviper

import (
	afero "github.com/spf13/afero"
	mapstructure "github.com/mitchellh/mapstructure"
	fsnotify "github.com/fsnotify/fsnotify"
	pflag "github.com/spf13/pflag"
	strings "strings"
	os "os"
	sync "sync"
	viper "github.com/spf13/viper"
	time "time"
	io "io"
)

var l sync.Mutex

func DecodeHook(hook mapstructure.DecodeHookFunc) viper.DecoderConfigOption {
	l.Lock()
	defer l.Unlock()
	return viper.DecodeHook(hook)
}

func New() *viper.Viper { l.Lock(); defer l.Unlock(); return viper.New() }

func KeyDelimiter(d string) viper.Option { l.Lock(); defer l.Unlock(); return viper.KeyDelimiter(d) }

func EnvKeyReplacer(r viper.StringReplacer) viper.Option {
	l.Lock()
	defer l.Unlock()
	return viper.EnvKeyReplacer(r)
}

func NewWithOptions(opts ...viper.Option) *viper.Viper {
	l.Lock()
	defer l.Unlock()
	return viper.NewWithOptions(opts...)
}

func Reset() { l.Lock(); defer l.Unlock(); viper.Reset() }

func OnConfigChange(run func(in fsnotify.Event)) {
	l.Lock()
	defer l.Unlock()
	viper.OnConfigChange(run)
}

func WatchConfig() { l.Lock(); defer l.Unlock(); viper.WatchConfig() }

func SetConfigFile(in string) { l.Lock(); defer l.Unlock(); viper.SetConfigFile(in) }

func SetEnvPrefix(in string) { l.Lock(); defer l.Unlock(); viper.SetEnvPrefix(in) }

func AllowEmptyEnv(allowEmptyEnv bool) {
	l.Lock()
	defer l.Unlock()
	viper.AllowEmptyEnv(allowEmptyEnv)
}

func ConfigFileUsed() string { l.Lock(); defer l.Unlock(); return viper.ConfigFileUsed() }

func AddConfigPath(in string) { l.Lock(); defer l.Unlock(); viper.AddConfigPath(in) }

func AddRemoteProvider(provider, endpoint, path string) error {
	l.Lock()
	defer l.Unlock()
	return viper.AddRemoteProvider(provider, endpoint, path)
}

func AddSecureRemoteProvider(provider, endpoint, path, secretkeyring string) error {
	l.Lock()
	defer l.Unlock()
	return viper.AddSecureRemoteProvider(provider, endpoint, path, secretkeyring)
}

func SetTypeByDefaultValue(enable bool) {
	l.Lock()
	defer l.Unlock()
	viper.SetTypeByDefaultValue(enable)
}

func GetViper() *viper.Viper { l.Lock(); defer l.Unlock(); return viper.GetViper() }

func Get(key string) interface{} { l.Lock(); defer l.Unlock(); return viper.Get(key) }

func Sub(key string) *viper.Viper { l.Lock(); defer l.Unlock(); return viper.Sub(key) }

func GetString(key string) string { l.Lock(); defer l.Unlock(); return viper.GetString(key) }

func GetBool(key string) bool { l.Lock(); defer l.Unlock(); return viper.GetBool(key) }

func GetInt(key string) int { l.Lock(); defer l.Unlock(); return viper.GetInt(key) }

func GetInt32(key string) int32 { l.Lock(); defer l.Unlock(); return viper.GetInt32(key) }

func GetInt64(key string) int64 { l.Lock(); defer l.Unlock(); return viper.GetInt64(key) }

func GetUint(key string) uint { l.Lock(); defer l.Unlock(); return viper.GetUint(key) }

func GetUint32(key string) uint32 { l.Lock(); defer l.Unlock(); return viper.GetUint32(key) }

func GetUint64(key string) uint64 { l.Lock(); defer l.Unlock(); return viper.GetUint64(key) }

func GetFloat64(key string) float64 { l.Lock(); defer l.Unlock(); return viper.GetFloat64(key) }

func GetTime(key string) time.Time { l.Lock(); defer l.Unlock(); return viper.GetTime(key) }

func GetDuration(key string) time.Duration { l.Lock(); defer l.Unlock(); return viper.GetDuration(key) }

func GetIntSlice(key string) []int { l.Lock(); defer l.Unlock(); return viper.GetIntSlice(key) }

func GetStringSlice(key string) []string {
	l.Lock()
	defer l.Unlock()
	return viper.GetStringSlice(key)
}

func GetStringMap(key string) map[string]interface{} {
	l.Lock()
	defer l.Unlock()
	return viper.GetStringMap(key)
}

func GetStringMapString(key string) map[string]string {
	l.Lock()
	defer l.Unlock()
	return viper.GetStringMapString(key)
}

func GetStringMapStringSlice(key string) map[string][]string {
	l.Lock()
	defer l.Unlock()
	return viper.GetStringMapStringSlice(key)
}

func GetSizeInBytes(key string) uint { l.Lock(); defer l.Unlock(); return viper.GetSizeInBytes(key) }

func UnmarshalKey(key string, rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	l.Lock()
	defer l.Unlock()
	return viper.UnmarshalKey(key, rawVal, opts...)
}

func Unmarshal(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	l.Lock()
	defer l.Unlock()
	return viper.Unmarshal(rawVal, opts...)
}

func UnmarshalExact(rawVal interface{}, opts ...viper.DecoderConfigOption) error {
	l.Lock()
	defer l.Unlock()
	return viper.UnmarshalExact(rawVal, opts...)
}

func BindPFlags(flags *pflag.FlagSet) error {
	l.Lock()
	defer l.Unlock()
	return viper.BindPFlags(flags)
}

func BindPFlag(key string, flag *pflag.Flag) error {
	l.Lock()
	defer l.Unlock()
	return viper.BindPFlag(key, flag)
}

func BindFlagValues(flags viper.FlagValueSet) error {
	l.Lock()
	defer l.Unlock()
	return viper.BindFlagValues(flags)
}

func BindFlagValue(key string, flag viper.FlagValue) error {
	l.Lock()
	defer l.Unlock()
	return viper.BindFlagValue(key, flag)
}

func BindEnv(input ...string) error { l.Lock(); defer l.Unlock(); return viper.BindEnv(input...) }

func IsSet(key string) bool { l.Lock(); defer l.Unlock(); return viper.IsSet(key) }

func AutomaticEnv() { l.Lock(); defer l.Unlock(); viper.AutomaticEnv() }

func SetEnvKeyReplacer(r *strings.Replacer) { l.Lock(); defer l.Unlock(); viper.SetEnvKeyReplacer(r) }

func RegisterAlias(alias string, key string) {
	l.Lock()
	defer l.Unlock()
	viper.RegisterAlias(alias, key)
}

func InConfig(key string) bool { l.Lock(); defer l.Unlock(); return viper.InConfig(key) }

func SetDefault(key string, value interface{}) {
	l.Lock()
	defer l.Unlock()
	viper.SetDefault(key, value)
}

func Set(key string, value interface{}) { l.Lock(); defer l.Unlock(); viper.Set(key, value) }

func ReadInConfig() error { l.Lock(); defer l.Unlock(); return viper.ReadInConfig() }

func MergeInConfig() error { l.Lock(); defer l.Unlock(); return viper.MergeInConfig() }

func ReadConfig(in io.Reader) error { l.Lock(); defer l.Unlock(); return viper.ReadConfig(in) }

func MergeConfig(in io.Reader) error { l.Lock(); defer l.Unlock(); return viper.MergeConfig(in) }

func MergeConfigMap(cfg map[string]interface{}) error {
	l.Lock()
	defer l.Unlock()
	return viper.MergeConfigMap(cfg)
}

func WriteConfig() error { l.Lock(); defer l.Unlock(); return viper.WriteConfig() }

func SafeWriteConfig() error { l.Lock(); defer l.Unlock(); return viper.SafeWriteConfig() }

func WriteConfigAs(filename string) error {
	l.Lock()
	defer l.Unlock()
	return viper.WriteConfigAs(filename)
}

func SafeWriteConfigAs(filename string) error {
	l.Lock()
	defer l.Unlock()
	return viper.SafeWriteConfigAs(filename)
}

func ReadRemoteConfig() error { l.Lock(); defer l.Unlock(); return viper.ReadRemoteConfig() }

func WatchRemoteConfig() error { l.Lock(); defer l.Unlock(); return viper.WatchRemoteConfig() }

func AllKeys() []string { l.Lock(); defer l.Unlock(); return viper.AllKeys() }

func AllSettings() map[string]interface{} { l.Lock(); defer l.Unlock(); return viper.AllSettings() }

func SetFs(fs afero.Fs) { l.Lock(); defer l.Unlock(); viper.SetFs(fs) }

func SetConfigName(in string) { l.Lock(); defer l.Unlock(); viper.SetConfigName(in) }

func SetConfigType(in string) { l.Lock(); defer l.Unlock(); viper.SetConfigType(in) }

func SetConfigPermissions(perm os.FileMode) {
	l.Lock()
	defer l.Unlock()
	viper.SetConfigPermissions(perm)
}

func Debug() { l.Lock(); defer l.Unlock(); viper.Debug() }
