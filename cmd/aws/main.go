package main

import (
	"context"
	"fmt"
	"os"
	"remindme/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

const (
	// This address must be verified with Amazon SES.
	Sender = "no-reply@remindme.one"
)

func main() {}

func CreateEmailTemplate(
	name string,
	subject string,
	htmlPart string,
	textPart string,
) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(
		context.Background(),
		awsConfig.WithRegion(cfg.AwsRegion),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AwsAccessKey,
				cfg.AwsSecretKey,
				"",
			),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	svc := ses.NewFromConfig(awsCfg)

	input := &ses.CreateTemplateInput{
		Template: &types.Template{
			SubjectPart:  &subject,
			HtmlPart:     &htmlPart,
			TextPart:     &textPart,
			TemplateName: &name,
		},
	}
	result, err := svc.CreateTemplate(context.Background(), input)

	// Display error messages if they occur.
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
		return
	}

	fmt.Println("Success:")
	fmt.Println(result)
}

func DeleteEmailTemplate(name string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(
		context.Background(),
		awsConfig.WithRegion(cfg.AwsRegion),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AwsAccessKey,
				cfg.AwsSecretKey,
				"",
			),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	svc := ses.NewFromConfig(awsCfg)

	result, err := svc.DeleteTemplate(
		context.Background(),
		&ses.DeleteTemplateInput{
			TemplateName: &name,
		},
	)

	// Display error messages if they occur.
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
		return
	}

	fmt.Println("Success:")
	fmt.Println(result)
}

func SendEmailTemplate(to string, name string, args string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(
		context.Background(),
		awsConfig.WithRegion(cfg.AwsRegion),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AwsAccessKey,
				cfg.AwsSecretKey,
				"",
			),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	svc := ses.NewFromConfig(awsCfg)

	result, err := svc.SendTemplatedEmail(
		context.Background(),
		&ses.SendTemplatedEmailInput{
			Source: aws.String(Sender),
			Destination: &types.Destination{
				CcAddresses: []string{},
				ToAddresses: []string{to},
			},
			Template:     &name,
			TemplateData: &args,
		},
	)

	// Display error messages if they occur.
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
		return
	}

	fmt.Println("Success:")
	fmt.Println(result)
}
