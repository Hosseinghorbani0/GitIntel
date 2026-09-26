package llm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type ProviderDiagnostic struct {
	HTTPStatus        int           `json:"http_status"`
	Classification    ErrorCategory `json:"classification"`
	ContentType       string        `json:"content_type"`
	ResponseBodyBytes int           `json:"response_body_bytes"`
	ProviderType      string        `json:"provider_type,omitempty"`
	ProviderCode      string        `json:"provider_code,omitempty"`
	ProviderMessage   string        `json:"provider_message,omitempty"`
}

type ProviderError struct {
	Category   ErrorCategory
	Message    string
	Status     int
	Cause      error
	Diagnostic *ProviderDiagnostic
}

func (e *ProviderError) Error() string {
	if e == nil {
		return "unknown llm error"
	}
	if e.Diagnostic != nil {
		diagnostic := e.Diagnostic
		return fmt.Sprintf("%s (HTTP %d, %s, content_type=%q, response_body_bytes=%d, provider_type=%q, provider_code=%q, provider_message=%q)",
			e.Message,
			diagnostic.HTTPStatus,
			diagnostic.Classification,
			diagnostic.ContentType,
			diagnostic.ResponseBodyBytes,
			diagnostic.ProviderType,
			diagnostic.ProviderCode,
			diagnostic.ProviderMessage,
		)
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *ProviderError) Unwrap() error { return e.Cause }

func classifyHTTPError(status int, body string) error {
	switch status {
	case http.StatusUnauthorized:
		return &ProviderError{Category: ErrorInvalidCredential, Message: "Invalid credential", Status: status}
	case http.StatusTooManyRequests:
		return &ProviderError{Category: ErrorRateLimited, Message: "Rate limit reached", Status: status}
	case http.StatusPaymentRequired:
		return &ProviderError{Category: ErrorPaymentRequired, Message: "Provider requires payment or available credits", Status: status}
	case http.StatusForbidden:
		return &ProviderError{Category: ErrorUnauthorized, Message: "Unauthorized", Status: status}
	case http.StatusServiceUnavailable:
		return &ProviderError{Category: ErrorProviderUnavailable, Message: "Provider unavailable", Status: status}
	case http.StatusNotFound:
		return &ProviderError{Category: ErrorModelUnavailable, Message: "Model unavailable", Status: status}
	case http.StatusRequestTimeout:
		return &ProviderError{Category: ErrorTimeout, Message: "Request timed out", Status: status}
	default:
		if status >= 500 {
			return &ProviderError{Category: ErrorProviderUnavailable, Message: "Provider returned a server error", Status: status}
		}
		if strings.Contains(strings.ToLower(body), "rate limit") || strings.Contains(strings.ToLower(body), "quota") {
			return &ProviderError{Category: ErrorQuotaExceeded, Message: "Quota or rate limit reached", Status: status}
		}
		if strings.Contains(strings.ToLower(body), "invalid") && strings.Contains(strings.ToLower(body), "token") {
			return &ProviderError{Category: ErrorInvalidCredential, Message: "Invalid credential", Status: status}
		}
		return &ProviderError{Category: ErrorUnknownProvider, Message: "Provider error", Status: status}
	}
}

func withProviderDiagnostic(err error, status int, contentType string, body []byte, credential string) error {
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		return err
	}
	var payload struct {
		Error   json.RawMessage `json:"error"`
		Type    string          `json:"type"`
		Code    string          `json:"code"`
		Message string          `json:"message"`
	}
	_ = json.Unmarshal(body, &payload)
	var providerFields struct {
		Type    string `json:"type"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if len(payload.Error) > 0 {
		if json.Unmarshal(payload.Error, &providerFields) != nil {
			_ = json.Unmarshal(payload.Error, &providerFields.Message)
		}
	}
	if providerFields.Type == "" {
		providerFields.Type = payload.Type
	}
	if providerFields.Code == "" {
		providerFields.Code = payload.Code
	}
	if providerFields.Message == "" {
		providerFields.Message = payload.Message
	}
	providerErr.Diagnostic = &ProviderDiagnostic{
		HTTPStatus:        status,
		Classification:    providerErr.Category,
		ContentType:       safeDiagnosticText(contentType, credential),
		ResponseBodyBytes: len(body),
		ProviderType:      safeDiagnosticText(providerFields.Type, credential),
		ProviderCode:      safeDiagnosticText(providerFields.Code, credential),
		ProviderMessage:   safeDiagnosticText(providerFields.Message, credential),
	}
	return providerErr
}

var secretLikeText = regexp.MustCompile(`(?i)(hf_[A-Za-z0-9_-]{12,}|Bearer\s+[A-Za-z0-9._~+/-]+=*)`)

func safeDiagnosticText(value, credential string) string {
	value = strings.TrimSpace(value)
	if credential != "" {
		value = strings.ReplaceAll(value, credential, "[redacted]")
	}
	value = secretLikeText.ReplaceAllString(value, "[redacted]")
	value = strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\t' {
			return ' '
		}
		return r
	}, value)
	if len(value) > 240 {
		value = value[:240]
	}
	return value
}

func IsInvalidCredential(err error) bool {
	var pe *ProviderError
	return errors.As(err, &pe) && pe.Category == ErrorInvalidCredential
}

func IsRateLimited(err error) bool {
	var pe *ProviderError
	return errors.As(err, &pe) && (pe.Category == ErrorRateLimited || pe.Category == ErrorQuotaExceeded)
}

func IsPaymentRequired(err error) bool {
	var pe *ProviderError
	return errors.As(err, &pe) && pe.Category == ErrorPaymentRequired
}

func MaskSecret(value string) string {
	if value == "" {
		return "not configured"
	}
	if strings.HasPrefix(value, "hf_") {
		if len(value) <= 12 {
			return "hf_***********" + value[len(value)-4:]
		}
		return value[:3] + "********" + value[len(value)-4:]
	}
	if len(value) <= 8 {
		return "********"
	}
	if len(value) <= 12 {
		return value[:3] + "********"
	}
	return value[:3] + "********" + value[len(value)-3:]
}
