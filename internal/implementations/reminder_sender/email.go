package remindersender

import (
	"context"
	"encoding/json"
	"fmt"
	"remindme/internal/core/domain/channel"
	"remindme/internal/core/domain/reminder"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type templateParams struct {
	ReminderBody string `json:"reminderBody"`
}

type EmailSender struct {
	ses *ses.Client
	// This address must be verified with Amazon SES.
	sender       string
	templateName string
}

func NewEmail(
	awsConfig aws.Config,
	sender string,
	templateName string,
) *EmailSender {
	return &EmailSender{
		ses:          ses.NewFromConfig(awsConfig),
		sender:       sender,
		templateName: templateName,
	}
}

func (s *EmailSender) SendReminder(
	ctx context.Context,
	rem reminder.Reminder,
	settings *channel.EmailSettings,
) error {
	body := rem.Body
	if body == "" {
		body = fmt.Sprintf("#%d", rem.ID)
	}
	templateArgsBytes, err := json.Marshal(templateParams{ReminderBody: body})
	if err != nil {
		return err
	}
	templateArgs := string(templateArgsBytes)

	email := string(settings.Email)
	_, err = s.ses.SendTemplatedEmail(
		ctx,
		&ses.SendTemplatedEmailInput{
			Source: aws.String(s.sender),
			Destination: &types.Destination{
				CcAddresses: []string{},
				ToAddresses: []string{email},
			},
			Template:     &s.templateName,
			TemplateData: &templateArgs,
		},
	)
	return err
}
