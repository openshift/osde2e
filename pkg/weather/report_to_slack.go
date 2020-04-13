package weather

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/report"
	"github.com/openshift/osde2e/pkg/common/templates"
	"github.com/slack-go/slack"
)

var (
	slackSummaryTemplate *template.Template
	slackJobTemplate     *template.Template
)

func init() {
	var err error
	slackSummaryTemplate, err = templates.LoadTemplate("/assets/reports/slack-summary.template")

	if err != nil {
		panic(fmt.Sprintf("error loading slack summary template: %v", err))
	}

	slackJobTemplate, err = templates.LoadTemplate("/assets/reports/slack-job.template")

	if err != nil {
		panic(fmt.Sprintf("error loading slack job template: %v", err))
	}
}

// SendReportToSlack will send the weather report to slack
func SendReportToSlack() error {
	if config.Instance.Weather.SlackWebhook == "" {
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

	jobAttachments, err := makeJobAttachments(report)

	if err != nil {
		return fmt.Errorf("error while making Slack job attachments: %v", err)
	}

	msg := &slack.WebhookMessage{
		Text:        "*osde2e weather report*",
		Attachments: append([]slack.Attachment{summaryAttachment}, jobAttachments...),
	}
	return slack.PostWebhook(config.Instance.Weather.SlackWebhook, msg)
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

func makeJobAttachments(w report.WeatherReport) ([]slack.Attachment, error) {
	slackJobAttachments := []slack.Attachment{}

	for _, job := range w.Jobs {
		slackJobBuffer := new(bytes.Buffer)

		if err := slackJobTemplate.ExecuteTemplate(slackJobBuffer, slackJobTemplate.Name(), job); err != nil {
			return nil, fmt.Errorf("error while creating slack job attachment: %v", err)
		}

		slackJobAttachments = append(slackJobAttachments, slack.Attachment{
			Pretext: fmt.Sprintf("*%s*", job.Name),
			Text:    string(slackJobBuffer.Bytes()),
		})
	}

	return slackJobAttachments, nil
}
