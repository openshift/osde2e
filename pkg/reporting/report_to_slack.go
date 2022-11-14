package reporting

import (
	"fmt"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/slack-go/slack"
)

// SendReportToSlack will send the weather report to slack
func SendReportToSlack(title string, report []byte) error {
	slackWebhook := viper.GetString(SlackWebhook)
	if slackWebhook == "" {
		return fmt.Errorf("no slack webhook configured")
	}

	summaryAttachment := slack.Attachment{
		Pretext: "*Summary*",
		Text:    string(report),
	}

	msg := &slack.WebhookMessage{
		Text:        title,
		Attachments: append([]slack.Attachment{summaryAttachment}),
	}
	return slack.PostWebhook(slackWebhook, msg)
}
