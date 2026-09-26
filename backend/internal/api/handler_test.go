package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitintel/backend/internal/analytics"
	"gitintel/backend/internal/cache"
	gh "gitintel/backend/internal/github"
	"gitintel/backend/internal/llm"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGitHubErrorStatus(t *testing.T) {
	tests := []struct {
		message string
		status  int
	}{
		{"404 Not Found", http.StatusNotFound},
		{"401 Bad credentials", http.StatusUnauthorized},
		{"403 Forbidden", http.StatusForbidden},
		{"403 API rate limit exceeded", http.StatusTooManyRequests},
		{"429 rate limit exceeded", http.StatusTooManyRequests},
		{"unexpected upstream error", http.StatusBadGateway},
	}
	for _, test := range tests {
		if got := githubErrorStatus(errors.New(test.message)); got != test.status {
			t.Errorf("githubErrorStatus(%q) = %d, want %d", test.message, got, test.status)
		}
	}
}

func TestInterpretIsDisabledWithoutBreakingDeterministicReport(t *testing.T) {
	t.Setenv("LLM_ANALYSIS_ENABLED", "false")
	cacheStore := cache.NewCache(time.Minute)
	analysis := analytics.Analysis{Summary: "Deterministic summary", ReportMarkdown: "Deterministic report"}
	evidence := llm.EvidenceContext{SchemaVersion: "1", Profile: llm.EvidenceProfile{Username: "octocat"}}
	cacheStore.Set("analysis:octocat", analyzeResult{Profile: gh.UserProfile{Username: "octocat"}, Analysis: analysis, Evidence: evidence})
	handler := NewHandler(cacheStore, analytics.NewEngine())

	response := httptest.NewRecorder()
	handler.Interpret(response, httptest.NewRequest(http.MethodPost, "/api/interpret", strings.NewReader(`{"username":"octocat","language":"en"}`)))
	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "LLM_ANALYSIS_DISABLED") {
		t.Fatalf("expected AI-disabled response, status=%d body=%s", response.Code, response.Body.String())
	}

	report := httptest.NewRecorder()
	handler.Report(report, httptest.NewRequest(http.MethodGet, "/api/report/octocat", nil))
	if report.Code != http.StatusOK || !strings.Contains(report.Body.String(), "Deterministic report") {
		t.Fatal("deterministic report must remain available when AI is disabled")
	}
}

func TestInterpretUsesSameCachedEvidenceForEnglishAndPersian(t *testing.T) {
	t.Setenv("LLM_ANALYSIS_ENABLED", "true")
	t.Setenv("LLM_PROVIDER", "huggingface")
	t.Setenv("HF_API_KEYS", "mock-test-key")
	t.Setenv("HF_API_KEY", "")
	t.Setenv("HF_MODEL", "mock-model")

	var userMessages []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode provider request: %v", err)
		}
		for _, message := range request.Messages {
			if message.Role == "user" {
				userMessages = append(userMessages, message.Content)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"choices":[{"message":{"content":"Interpretation"}}],"usage":{"total_tokens":8}}`)
	}))
	defer server.Close()
	t.Setenv("HF_BASE_URL", server.URL)

	evidence := llm.EvidenceContext{
		SchemaVersion:           "1",
		RepositoryDataUntrusted: true,
		Profile:                 llm.EvidenceProfile{Username: "octocat"},
		Signals:                 []llm.EvidenceSignal{{Name: "Maintenance", Classification: "Moderate", Metric: map[string]float64{"updated_180d": 2}, Observations: []string{"2 updated"}, Limitations: []string{"timestamps only"}}},
	}
	cacheStore := cache.NewCache(time.Minute)
	cacheStore.Set("analysis:octocat", analyzeResult{Profile: gh.UserProfile{Username: "octocat"}, Evidence: evidence})
	handler := NewHandler(cacheStore, analytics.NewEngine())
	for _, language := range []string{"en", "fa"} {
		body := fmt.Sprintf(`{"username":"octocat","language":%q}`, language)
		response := httptest.NewRecorder()
		handler.Interpret(response, httptest.NewRequest(http.MethodPost, "/api/interpret", strings.NewReader(body)))
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"language":"`+language+`"`) {
			t.Fatalf("interpretation for %s failed: status=%d body=%s", language, response.Code, response.Body.String())
		}
	}
	if len(userMessages) != 2 {
		t.Fatalf("expected two mock provider requests, got %d", len(userMessages))
	}
	for _, message := range userMessages {
		if !strings.Contains(message, `"username":"octocat"`) || !strings.Contains(message, `"classification":"Moderate"`) {
			t.Fatal("provider request did not contain cached deterministic evidence")
		}
	}
	if !strings.Contains(userMessages[0], "Interpret the supplied structured GitHub evidence") || !strings.Contains(userMessages[1], "شواهد ساخت‌یافته") {
		t.Fatal("expected language-specific instructions with shared evidence")
	}
}

func TestInterpretPaymentFailurePreservesDeterministicReport(t *testing.T) {
	t.Setenv("LLM_ANALYSIS_ENABLED", "true")
	t.Setenv("LLM_PROVIDER", "huggingface")
	t.Setenv("HF_API_KEYS", "mock-test-key")
	t.Setenv("HF_API_KEY", "")
	t.Setenv("HF_MODEL", "mock-model")
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		_, _ = fmt.Fprint(w, `{"error":"Payment required"}`)
	}))
	defer server.Close()
	t.Setenv("HF_BASE_URL", server.URL)

	cacheStore := cache.NewCache(time.Minute)
	cacheStore.Set("analysis:octocat", analyzeResult{
		Profile:  gh.UserProfile{Username: "octocat"},
		Analysis: analytics.Analysis{ReportMarkdown: "Deterministic fallback report"},
		Evidence: llm.EvidenceContext{SchemaVersion: "1", Profile: llm.EvidenceProfile{Username: "octocat"}},
	})
	handler := NewHandler(cacheStore, analytics.NewEngine())
	interpretation := httptest.NewRecorder()
	handler.Interpret(interpretation, httptest.NewRequest(http.MethodPost, "/api/interpret", strings.NewReader(`{"username":"octocat","language":"en"}`)))
	if interpretation.Code != http.StatusPaymentRequired || requests != 1 {
		t.Fatalf("expected one payment-required provider request, status=%d requests=%d", interpretation.Code, requests)
	}
	report := httptest.NewRecorder()
	handler.Report(report, httptest.NewRequest(http.MethodGet, "/api/report/octocat", nil))
	if report.Code != http.StatusOK || !strings.Contains(report.Body.String(), "Deterministic fallback report") {
		t.Fatal("deterministic report must remain available after AI provider failure")
	}
}
