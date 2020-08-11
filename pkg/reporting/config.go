package reporting

import "github.com/spf13/viper"

const (
	// SlackWebhook for pushing reports to slack
	SlackWebhook = "reporting.slackWebhook"
)

func init() {
	viper.BindEnv(SlackWebhook, "SLACK_WEBHOOK")
}
