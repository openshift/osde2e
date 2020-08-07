package weather

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/report"
	"github.com/openshift/osde2e/pkg/common/templates"
	"github.com/slack-go/slack"
	"github.com/spf13/viper"
)

var (
	slackSummaryTemplate *template.Template
)

func init() {
	var err error
	slackSummaryTemplate, err = templates.LoadTemplate("/assets/reports/slack-summary.template")

	if err != nil {
		panic(fmt.Sprintf("error loading slack summary template: %v", err))
	}
}

// SendReportToSlack will send the weather report to slack
func SendReportToSlack() error {
	slackWebhook := viper.GetString(config.Weather.SlackWebhook)
	if slackWebhook == "" {
		return fmt.Errorf("no slack webhook configured")
	}

	report, err := report.GenerateReport()

	if err != nil {
		return fmt.Errorf("error while generating report: %v", err)
	}

	summaryAttachment, err := makeSummaryAttachment(report)

	if err != nil {
		return fmt.Errorf("error while making Slack summary attachment: %v", err)
	}

	msg := &slack.WebhookMessage{
		Text:        "*osde2e weather report*",
		Attachments: append([]slack.Attachment{summaryAttachment}),
	}
	return slack.PostWebhook(slackWebhook, msg)
}

func makeSummaryAttachment(w report.WeatherReport) (slack.Attachment, error) {
	slackSummaryBuffer := new(bytes.Buffer)

	if err := slackSummaryTemplate.ExecuteTemplate(slackSummaryBuffer, slackSummaryTemplate.Name(), w); err != nil {
		return slack.Attachment{}, fmt.Errorf("error while creating slack summary attachment: %v", err)
	}

	return slack.Attachment{
		Pretext: "*Summary*",
		Text:    string(slackSummaryBuffer.Bytes()),
	}, nil
}
