package weather

import viper "github.com/openshift/osde2e/pkg/common/concurrentviper"

const (
	StartOfTimeWindowInHours = "reporting.weather.startOfTimeWindowInHours"
	NumberOfSamplesNecessary = "reporting.weather..numberOfSamplesNecessary"
	JobAllowlist             = "reporting.weather.jobAllowlist"
	Provider                 = "weather.provider"
)

func init() {
	viper.SetDefault(StartOfTimeWindowInHours, 24)
	viper.BindEnv(StartOfTimeWindowInHours, "REPORTING_WEATHER_START_OF_TIME_WINDOW_IN_HOURS")

	viper.SetDefault(NumberOfSamplesNecessary, 3)
	viper.BindEnv(NumberOfSamplesNecessary, "REPORTING_WEATHER_NUMBER_OF_SAMPLES_NECESSARY")

	viper.SetDefault(JobAllowlist, "osde2e-.*-aws-e2e-.*")
	viper.BindEnv(JobAllowlist, "REPORTING_WEATHER_JOB_ALLOWLIST")

	viper.SetDefault(Provider, "aws")
	viper.BindEnv(Provider, "REPORTING_WEATHER_PROVIDER")
}
