package email

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"

	"remindme/internal/core/domain/channel"
	"remindme/internal/core/domain/user"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type EmailSender struct {
	ses *ses.Client
	// This address must be verified with Amazon SES.
	sender                    string
	accountActivationTemplate string
	accountActivationUrl      url.URL
	passwordResetTemplate     string
	passwordResetBaseUrl      url.URL
	channelActivationTemplate string
}

func NewEmailSender(
	awsConfig aws.Config,
	sender string,
	accountActivationTemplate string,
	accountActivationUrl url.URL,
	passwordResetTemplate string,
	passwordResetBaseUrl url.URL,
	channelActivationTemplate string,
) *EmailSender {
	return &EmailSender{
		ses:                       ses.NewFromConfig(awsConfig),
		sender:                    sender,
		accountActivationTemplate: accountActivationTemplate,
		accountActivationUrl:      accountActivationUrl,
		passwordResetTemplate:     passwordResetTemplate,
		passwordResetBaseUrl:      passwordResetBaseUrl,
		channelActivationTemplate: channelActivationTemplate,
	}
}

func (s *EmailSender) SendActivationToken(ctx context.Context, u user.User) error {
	if !u.ActivationToken.IsPresent {
		return errors.New("user activation token is not defined")
	}
	if !u.Email.IsPresent {
		return errors.New("user email is not defined")
	}

	templateParamsBytes, err := json.Marshal(
		accountActivationTemplateParams{
			ActivationCode: string(u.ActivationToken.Value),
			ActivationUrl:  s.accountActivationUrl.String(),
		},
	)
	if err != nil {
		return err
	}
	templateParams := string(templateParamsBytes)

	email := string(u.Email.Value)
	_, err = s.ses.SendTemplatedEmail(
		ctx,
		&ses.SendTemplatedEmailInput{
			Source: &s.sender,
			Destination: &types.Destination{
				CcAddresses: []string{},
				ToAddresses: []string{email},
			},
			Template:     &s.accountActivationTemplate,
			TemplateData: &templateParams,
		},
	)
	return err
}

func (s *EmailSender) SendPasswordResetToken(ctx context.Context, u user.User, token user.PasswordResetToken) error {
	if !u.Email.IsPresent {
		return errors.New("user email is not defined")
	}

	templateParamsBytes, err := json.Marshal(
		passwordResetTemplateParams{
			PasswordResetUrl: s.passwordResetBaseUrl.JoinPath(string(token)).String(),
		},
	)
	if err != nil {
		return err
	}
	templateParams := string(templateParamsBytes)

	email := string(u.Email.Value)
	_, err = s.ses.SendTemplatedEmail(
		ctx,
		&ses.SendTemplatedEmailInput{
			Source: &s.sender,
			Destination: &types.Destination{
				CcAddresses: []string{},
				ToAddresses: []string{email},
			},
			Template:     &s.passwordResetTemplate,
			TemplateData: &templateParams,
		},
	)
	return err
}

func (s *EmailSender) SendVerificationToken(
	ctx context.Context,
	token channel.VerificationToken,
	c channel.Channel,
) error {
	settings, ok := c.Settings.(*channel.EmailSettings)
	if !ok {
		return errors.New("not email channel")
	}

	templateParamsBytes, err := json.Marshal(
		channelActivationTemplateParams{
			ActivationCode: string(token),
		},
	)
	if err != nil {
		return err
	}
	templateParams := string(templateParamsBytes)

	email := string(settings.Email)
	_, err = s.ses.SendTemplatedEmail(
		ctx,
		&ses.SendTemplatedEmailInput{
			Source: &s.sender,
			Destination: &types.Destination{
				CcAddresses: []string{},
				ToAddresses: []string{email},
			},
			Template:     &s.channelActivationTemplate,
			TemplateData: &templateParams,
		},
	)
	return err
}

type accountActivationTemplateParams struct {
	ActivationCode string `json:"activationCode"`
	ActivationUrl  string `json:"activationUrl"`
}

type passwordResetTemplateParams struct {
	PasswordResetUrl string `json:"passwordResetUrl"`
}

type channelActivationTemplateParams struct {
	ActivationCode string `json:"activationCode"`
}
