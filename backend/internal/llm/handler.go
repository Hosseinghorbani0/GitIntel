package llm

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func HandleTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeLLMError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST is required.")
		return
	}
	var input struct {
		Language string `json:"language"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 16*1024)).Decode(&input); err != nil && !errors.Is(err, io.EOF) {
		writeLLMError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid model-test request.")
		return
	}
	language := strings.ToLower(strings.TrimSpace(input.Language))
	if language == "" {
		language = "en"
	}
	if language != "en" && language != "fa" {
		writeLLMError(w, http.StatusBadRequest, "INVALID_LANGUAGE", "Language must be en or fa.")
		return
	}
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		writeLLMError(w, http.StatusBadRequest, "INVALID_LLM_CONFIG", err.Error())
		return
	}
	if ProviderAvailabilityStatus(cfg) == "payment_required" {
		writeLLMProviderError(w, http.StatusPaymentRequired, "PAYMENT_REQUIRED", paymentRequiredProviderError())
		return
	}
	manager := NewCredentialManager(CredentialEnvName(cfg.Provider))
	manager.Load()
	if !manager.HasCredentials() {
		writeLLMError(w, http.StatusServiceUnavailable, "MISSING_LLM_CREDENTIALS", "No provider credential is configured.")
		return
	}
	for {
		credential, ok := manager.SelectAvailable()
		if !ok {
			break
		}
		cfg.Credential = credential.Value
		provider, err := BuildProvider(cfg)
		if err != nil {
			writeLLMError(w, http.StatusBadRequest, "INVALID_LLM_PROVIDER", err.Error())
			return
		}
		result, err := (&ModelTester{provider: provider, config: cfg}).TestContext(r.Context(), language)
		if err == nil {
			result.CredentialStatus = "active"
			writeLLMResponse(w, http.StatusOK, map[string]interface{}{"success": true, "data": result})
			return
		}
		if IsInvalidCredential(err) {
			manager.MarkInvalid(credential.Value, "invalid")
			continue
		}
		if IsPaymentRequired(err) {
			writeLLMProviderError(w, http.StatusPaymentRequired, "PAYMENT_REQUIRED", err)
			return
		}
		if providerErr, ok := err.(*ProviderError); ok && providerErr.Category == ErrorTimeout {
			writeLLMProviderError(w, http.StatusGatewayTimeout, "LLM_TIMEOUT", err)
			return
		}
		if IsRateLimited(err) {
			writeLLMProviderError(w, http.StatusTooManyRequests, "RATE_LIMITED", err)
			return
		}
		writeLLMProviderError(w, http.StatusBadGateway, "LLM_REQUEST_FAILED", err)
		return
	}
	writeLLMError(w, http.StatusUnauthorized, "INVALID_CREDENTIAL", "All configured Hugging Face credentials were rejected.")
}

func HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeLLMError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET is required.")
		return
	}
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		writeLLMError(w, http.StatusBadRequest, "INVALID_LLM_CONFIG", "The LLM configuration is invalid.")
		return
	}
	manager := NewCredentialManager(CredentialEnvName(cfg.Provider))
	credentialStatus := "not configured"
	if manager.HasCredentials() {
		credentialStatus = "configured"
	}
	providerStatus := ProviderAvailabilityStatus(cfg)
	writeLLMResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"provider":                 cfg.Provider,
			"model":                    cfg.Model,
			"credential_status":        credentialStatus,
			"provider_status":          providerStatus,
			"analysis_reports_enabled": AnalysisReportsEnabled(),
		},
	})
}

func writeLLMResponse(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeLLMProviderError(w http.ResponseWriter, status int, code string, err error) {
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		writeLLMError(w, status, code, "The model provider request failed.")
		return
	}
	writeLLMResponse(w, status, map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":       code,
			"message":    providerErr.Error(),
			"diagnostic": providerErr.Diagnostic,
		},
	})
}

func writeLLMError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   map[string]string{"code": code, "message": message},
	})
}
