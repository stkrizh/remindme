package recaptcha

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/services/captcha"
	"strings"
	"time"
)

const RECAPTCHA_VERIFICATION_URL = "https://www.google.com/recaptcha/api/siteverify"

type VerificationResult struct {
	Success  bool   `json:"success"`
	Hostname string `json:"hostname"`
}

func (r *VerificationResult) FromJSON(reader io.Reader) error {
	decoder := json.NewDecoder(reader)
	return decoder.Decode(r)
}

type GoogleRecaptchaValidator struct {
	log            logging.Logger
	httpClient     http.Client
	scoreThreshold float64
	secretKey      string
}

func New(
	log logging.Logger,
	secretKey string,
	scoreThreshold float64,
	timeout time.Duration,
) *GoogleRecaptchaValidator {
	if log == nil {
		panic(e.NewNilArgumentError("validator"))
	}
	return &GoogleRecaptchaValidator{
		log:            log,
		scoreThreshold: scoreThreshold,
		secretKey:      secretKey,
		httpClient:     http.Client{Timeout: timeout},
	}
}

func (v *GoogleRecaptchaValidator) ValidateCaptchaToken(ctx context.Context, token captcha.CaptchaToken) bool {
	if token.IsZero() {
		v.log.Info(ctx, "Recaptcha token is not provided, skip verification.")
		return false
	}

	requestBody := url.Values{}
	requestBody.Add("secret", v.secretKey)
	requestBody.Add("response", string(token))

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		RECAPTCHA_VERIFICATION_URL,
		strings.NewReader(requestBody.Encode()),
	)
	if err != nil {
		logging.Error(ctx, v.log, err)
		return true
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := v.httpClient.Do(request)
	if err != nil {
		logging.Error(ctx, v.log, err)
		return true
	}
	defer response.Body.Close()

	result := VerificationResult{}
	result.FromJSON(response.Body)
	v.log.Info(
		ctx,
		"Recaptcha token has been validated.",
		logging.Entry("result", result),
	)
	return result.Success
}
