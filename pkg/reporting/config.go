package reporting

import viper "github.com/openshift/osde2e/pkg/common/concurrentviper"

const (
	// SlackWebhook for pushing reports to slack
	SlackWebhook = "reporting.slackWebhook"
)

func init() {
	viper.BindEnv(SlackWebhook, "SLACK_WEBHOOK")
}
