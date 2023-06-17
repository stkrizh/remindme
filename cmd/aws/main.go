package main

import (
	"fmt"
	"os"
	"remindme/internal/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

const (
	// This address must be verified with Amazon SES.
	Sender = "no-reply@remindme.one"
)

func main() {
}

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

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.AwsRegion),
		Credentials: credentials.NewStaticCredentials(cfg.AwsAccessKey, cfg.AwsSecretKey, ""),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	svc := ses.New(sess)

	input := &ses.CreateTemplateInput{
		Template: &ses.Template{
			SubjectPart:  &subject,
			HtmlPart:     &htmlPart,
			TextPart:     &textPart,
			TemplateName: &name,
		},
	}
	result, err := svc.CreateTemplate(input)

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

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.AwsRegion),
		Credentials: credentials.NewStaticCredentials(cfg.AwsAccessKey, cfg.AwsSecretKey, ""),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	svc := ses.New(sess)

	result, err := svc.DeleteTemplate(&ses.DeleteTemplateInput{
		TemplateName: &name,
	})

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

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.AwsRegion),
		Credentials: credentials.NewStaticCredentials(cfg.AwsAccessKey, cfg.AwsSecretKey, ""),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	svc := ses.New(sess)

	result, err := svc.SendTemplatedEmail(&ses.SendTemplatedEmailInput{
		Source: aws.String(Sender),
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(to),
			},
		},
		Template:     &name,
		TemplateData: &args,
	})

	// Display error messages if they occur.
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
		return
	}

	fmt.Println("Success:")
	fmt.Println(result)
}
