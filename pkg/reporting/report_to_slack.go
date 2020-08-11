package reporting

import (
	"fmt"

	"github.com/slack-go/slack"
	"github.com/spf13/viper"
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
