package llm

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCredentialManagerMaskingAndSelection(t *testing.T) {
	t.Setenv("HF_API_KEYS", "hf_abc123,hf_def456")

	manager := NewCredentialManager("HF_API_KEYS")
	creds := manager.Load()
	if len(creds) != 2 {
		t.Fatalf("expected 2 credentials, got %d", len(creds))
	}
	if got := MaskSecret("hf_abc123"); got != "hf_***********c123" {
		t.Fatalf("unexpected masked value: %s", got)
	}
	if !manager.HasCredentials() {
		t.Fatal("expected configured credentials to be available")
	}
}

func TestConfigLoadUsesEnvironmentDefaults(t *testing.T) {
	t.Setenv("LLM_PROVIDER", "huggingface")
	t.Setenv("HF_BASE_URL", "https://router.huggingface.co/v1")
	t.Setenv("HF_MODEL", "Qwen/Qwen3.8-2.4T-A95B:novita")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected config to load: %v", err)
	}
	if cfg.Provider != "huggingface" {
		t.Fatalf("expected provider huggingface, got %s", cfg.Provider)
	}
	if cfg.Model == "" {
		t.Fatal("expected configured model")
	}
}

func TestHuggingFaceProviderSuccessAndRateLimitHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chat/completions" {
			if r.Header.Get("Authorization") == "Bearer test-key" {
				var payload struct {
					Model    string `json:"model"`
					Stream   bool   `json:"stream"`
					Messages []struct {
						Content string `json:"content"`
					} `json:"messages"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Errorf("expected valid OpenAI-compatible request body: %v", err)
				}
				if payload.Model != "test-model" || payload.Stream || len(payload.Messages) != 2 || !strings.Contains(payload.Messages[1].Content, `"repository":"example-project"`) {
					t.Errorf("expected non-streaming request to include deterministic evidence")
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"abc","choices":[{"message":{"content":"evidence backed summary"}}],"usage":{"prompt_tokens":10,"completion_tokens":20,"total_tokens":30}}`))
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"message":"invalid token"}}`))
			return
		}
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"rate limit exceeded"}}`))
	}))
	defer server.Close()

	cfg := Config{Provider: "huggingface", Model: "test-model", BaseURL: server.URL, Credential: "test-key", Temperature: 0.2, MaxTokens: 200}
	provider := NewHuggingFaceProvider(cfg)
	response, err := provider.Generate(TestPrompt("en"))
	if err != nil {
		t.Fatalf("expected generation success, got %v", err)
	}
	if response == nil || response.Text == "" {
		t.Fatal("expected non-empty model response")
	}

	cfg2 := Config{Provider: "huggingface", Model: "test-model", BaseURL: server.URL, Credential: "bad-key", Temperature: 0.2, MaxTokens: 200}
	provider2 := NewHuggingFaceProvider(cfg2)
	_, err = provider2.Generate(TestPrompt("en"))
	if err == nil {
		t.Fatal("expected unauthorized error for invalid credential")
	}
	if !IsInvalidCredential(err) {
		t.Fatalf("expected invalid credential classification, got %v", err)
	}
}

func TestTestPromptIncludesEvidenceGuardrails(t *testing.T) {
	prompt := TestPrompt("en")
	if prompt.System == "" || prompt.User == "" {
		t.Fatal("system and user prompt should both be populated")
	}
	if prompt.Input == nil {
		t.Fatal("test input should not be empty")
	}
}

func TestTestPromptSupportsEnglishAndPersian(t *testing.T) {
	english := TestPrompt("en")
	persian := TestPrompt("fa")
	if english.Language != "en" || !strings.Contains(english.System, "in English") {
		t.Fatal("English prompt must explicitly request English")
	}
	if persian.Language != "fa" || !strings.Contains(persian.System, "natural Persian") || !strings.Contains(persian.User, "گزارش") {
		t.Fatal("Persian prompt must request natural professional Persian")
	}
	if len(english.Input.(map[string]interface{})) != len(persian.Input.(map[string]interface{})) {
		t.Fatal("both languages must use the same deterministic evidence")
	}
}

func TestHandleTestRetriesOnlyInvalidCredentials(t *testing.T) {
	t.Setenv("HF_API_KEYS", "rejected-key,working-key")
	t.Setenv("HF_API_KEY", "")
	t.Setenv("LLM_PROVIDER", "huggingface")
	t.Setenv("HF_MODEL", "test-model")

	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Header.Get("Authorization") == "Bearer rejected-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"message":"invalid token"}}`))
			return
		}
		var payload struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if len(payload.Messages) != 2 || !strings.Contains(payload.Messages[0].Content, "natural Persian") {
			t.Errorf("expected Persian language instructions in provider request")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"پاسخ حرفه‌ای"}}],"usage":{"total_tokens":12}}`))
	}))
	defer server.Close()
	t.Setenv("HF_BASE_URL", server.URL)

	request := httptest.NewRequest(http.MethodPost, "/api/llm/test", strings.NewReader(`{"language":"fa"}`))
	response := httptest.NewRecorder()
	HandleTest(response, request)
	if response.Code != http.StatusOK || requests != 2 {
		t.Fatalf("expected success after one invalid credential, status=%d requests=%d", response.Code, requests)
	}
	body, _ := io.ReadAll(response.Body)
	if strings.Contains(string(body), "rejected-key") || strings.Contains(string(body), "working-key") {
		t.Fatal("API response exposed a configured credential")
	}
	if !strings.Contains(string(body), `"credential_status":"active"`) {
		t.Fatal("expected only the active credential state in response")
	}
}

func TestHandleTestDoesNotRotateOnRateLimit(t *testing.T) {
	t.Setenv("HF_API_KEYS", "first-key,second-key")
	t.Setenv("HF_API_KEY", "")
	t.Setenv("LLM_PROVIDER", "huggingface")
	t.Setenv("HF_MODEL", "test-model")
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"rate limit exceeded"}}`))
	}))
	defer server.Close()
	t.Setenv("HF_BASE_URL", server.URL)

	response := httptest.NewRecorder()
	HandleTest(response, httptest.NewRequest(http.MethodPost, "/api/llm/test", strings.NewReader(`{"language":"en"}`)))
	if response.Code != http.StatusTooManyRequests || requests != 1 {
		t.Fatalf("rate limit must stop after one request, status=%d requests=%d", response.Code, requests)
	}
}

func TestHandleTestDoesNotRotateOnPaymentRequired(t *testing.T) {
	t.Setenv("HF_API_KEYS", "first-test-key,second-test-key")
	t.Setenv("HF_API_KEY", "")
	t.Setenv("LLM_PROVIDER", "huggingface")
	t.Setenv("HF_MODEL", "test-model")
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = w.Write([]byte(`{"error":"Payment required"}`))
	}))
	defer server.Close()
	t.Setenv("HF_BASE_URL", server.URL)

	response := httptest.NewRecorder()
	HandleTest(response, httptest.NewRequest(http.MethodPost, "/api/llm/test", strings.NewReader(`{"language":"en"}`)))
	if response.Code != http.StatusPaymentRequired || requests != 1 {
		t.Fatalf("payment-required response must not rotate credentials, status=%d requests=%d", response.Code, requests)
	}
	if !strings.Contains(response.Body.String(), `"code":"PAYMENT_REQUIRED"`) {
		t.Fatal("expected distinct payment-required API error")
	}
	if !strings.Contains(response.Body.String(), `"provider_message":"Payment required"`) {
		t.Fatal("expected safe diagnostic extraction from a string-shaped provider error")
	}
}

func TestHandleStatusDoesNotExposeCredentials(t *testing.T) {
	t.Setenv("HF_API_KEYS", "private-test-key,another-private-key")
	response := httptest.NewRecorder()
	HandleStatus(response, httptest.NewRequest(http.MethodGet, "/api/llm/status", nil))
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "private-test-key") {
		t.Fatal("credential status endpoint exposed a credential")
	}
	if !strings.Contains(response.Body.String(), `"credential_status":"configured"`) {
		t.Fatal("expected configured credential state")
	}
}

func TestProviderDiagnosticsClassifyAndRedactUpstreamErrors(t *testing.T) {
	t.Setenv("HF_API_KEYS", "local-test-credential")
	t.Setenv("HF_API_KEY", "")
	t.Setenv("LLM_PROVIDER", "huggingface")
	t.Setenv("HF_MODEL", "test-model")
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":{"message":"upstream rejected local-test-credential","type":"upstream_error","code":"model_route_unavailable"}}`))
	}))
	defer server.Close()
	t.Setenv("HF_BASE_URL", server.URL)

	response := httptest.NewRecorder()
	HandleTest(response, httptest.NewRequest(http.MethodPost, "/api/llm/test", strings.NewReader(`{"language":"en"}`)))
	if response.Code != http.StatusBadGateway || requestCount != 1 {
		t.Fatalf("expected one upstream 502, got status=%d requests=%d", response.Code, requestCount)
	}
	body := response.Body.String()
	if strings.Contains(body, "local-test-credential") {
		t.Fatal("provider diagnostic response exposed the configured credential")
	}
	var payload struct {
		Error struct {
			Diagnostic ProviderDiagnostic `json:"diagnostic"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid diagnostic JSON: %v", err)
	}
	diagnostic := payload.Error.Diagnostic
	if diagnostic.HTTPStatus != http.StatusBadGateway || diagnostic.Classification != ErrorProviderUnavailable {
		t.Fatalf("unexpected upstream classification: %+v", diagnostic)
	}
	if diagnostic.ContentType != "application/json; charset=utf-8" || diagnostic.ResponseBodyBytes == 0 {
		t.Fatalf("expected safe content-type and body length, got %+v", diagnostic)
	}
	if diagnostic.ProviderType != "upstream_error" || diagnostic.ProviderCode != "model_route_unavailable" {
		t.Fatalf("expected allowlisted provider error fields, got %+v", diagnostic)
	}
	if strings.Contains(diagnostic.ProviderMessage, "local-test-credential") {
		t.Fatal("provider message exposed the configured credential")
	}
	if !strings.Contains(diagnostic.ProviderMessage, "[redacted]") {
		t.Fatal("expected echoed credential to be redacted from provider message")
	}
}

func TestProviderHTTPErrorCategories(t *testing.T) {
	tests := []struct {
		status   int
		category ErrorCategory
	}{
		{http.StatusUnauthorized, ErrorInvalidCredential},
		{http.StatusForbidden, ErrorUnauthorized},
		{http.StatusTooManyRequests, ErrorRateLimited},
		{http.StatusPaymentRequired, ErrorPaymentRequired},
		{http.StatusBadGateway, ErrorProviderUnavailable},
		{http.StatusServiceUnavailable, ErrorProviderUnavailable},
	}
	for _, test := range tests {
		err := classifyHTTPError(test.status, "")
		var providerErr *ProviderError
		if !errors.As(err, &providerErr) || providerErr.Category != test.category {
			t.Errorf("status %d: expected category %s, got %v", test.status, test.category, err)
		}
	}
}

func TestHuggingFaceProviderClassifiesMalformedAndEmptyResponses(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed JSON", body: "not-json"},
		{name: "empty completion", body: `{"choices":[]}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, test.body)
			}))
			defer server.Close()
			provider := NewHuggingFaceProvider(Config{Provider: "huggingface", Model: "test-model", BaseURL: server.URL, Timeout: 2})
			_, err := provider.Generate(TestPrompt("en"))
			var providerErr *ProviderError
			if !errors.As(err, &providerErr) || providerErr.Category != ErrorInvalidResponse {
				t.Fatalf("expected invalid-response classification, got %v", err)
			}
			if providerErr.Diagnostic == nil || providerErr.Diagnostic.HTTPStatus != http.StatusOK || providerErr.Diagnostic.ResponseBodyBytes != len(test.body) {
				t.Fatalf("expected safe response metadata, got %+v", providerErr.Diagnostic)
			}
		})
	}
}

func TestHuggingFaceProviderClassifiesTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	provider := NewHuggingFaceProvider(Config{Provider: "huggingface", Model: "test-model", BaseURL: server.URL, Timeout: 1})
	_, err := provider.Generate(TestPrompt("en"))
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) || providerErr.Category != ErrorTimeout {
		t.Fatalf("expected timeout classification, got %v", err)
	}
}
