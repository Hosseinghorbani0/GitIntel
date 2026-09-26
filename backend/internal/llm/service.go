package llm

import (
	"context"
	"errors"
	"os"
	"strings"
)

var ErrAnalysisDisabled = errors.New("AI report generation is disabled")

const LastKnownPaymentRequiredModel = "Qwen/Qwen3.8-2.4T-A95B:novita"

func ProviderAvailabilityStatus(cfg Config) string {
	configured := strings.ToLower(strings.TrimSpace(os.Getenv("LLM_PROVIDER_STATUS")))
	switch configured {
	case "available", "payment_required", "rate_limited", "unavailable", "unverified", "disabled":
		return configured
	}
	if strings.EqualFold(strings.TrimSpace(cfg.Provider), "huggingface") && strings.TrimSpace(cfg.Model) == LastKnownPaymentRequiredModel {
		return "payment_required"
	}
	if !AnalysisReportsEnabled() {
		return "disabled"
	}
	return "unverified"
}

func paymentRequiredProviderError() *ProviderError {
	return &ProviderError{
		Category: ErrorPaymentRequired,
		Message:  "Provider is blocked after a verified payment-required response",
		Status:   402,
		Diagnostic: &ProviderDiagnostic{
			HTTPStatus:        402,
			Classification:    ErrorPaymentRequired,
			ProviderMessage:   "Last verified provider response requires payment or available credits.",
			ResponseBodyBytes: 0,
		},
	}
}

func AnalysisReportsEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("LLM_ANALYSIS_ENABLED")), "true")
}

func GenerateInterpretation(ctx context.Context, cfg Config, evidence EvidenceContext, language string) (*Response, error) {
	if !AnalysisReportsEnabled() {
		return nil, ErrAnalysisDisabled
	}
	if ProviderAvailabilityStatus(cfg) == "payment_required" {
		return nil, paymentRequiredProviderError()
	}
	manager := NewCredentialManager(CredentialEnvName(cfg.Provider))
	manager.Load()
	for {
		credential, ok := manager.SelectAvailable()
		if !ok {
			return nil, &ProviderError{Category: ErrorInvalidCredential, Message: "No usable provider credential is configured"}
		}
		cfg.Credential = credential.Value
		provider, err := BuildProvider(cfg)
		if err != nil {
			return nil, err
		}
		request := ReportPrompt(evidence, language)
		var result *Response
		if contextualProvider, ok := provider.(interface {
			GenerateContext(context.Context, Request) (*Response, error)
		}); ok {
			result, err = contextualProvider.GenerateContext(ctx, request)
		} else {
			result, err = provider.Generate(request)
		}
		if err == nil {
			return result, nil
		}
		if IsInvalidCredential(err) {
			manager.MarkInvalid(credential.Value, "invalid")
			continue
		}
		return nil, err
	}
}

func CredentialEnvName(provider string) string {
	if strings.EqualFold(strings.TrimSpace(provider), "huggingface") {
		return "HF_API_KEYS"
	}
	return "LLM_API_KEYS"
}
