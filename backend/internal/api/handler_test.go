package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gitintel/backend/internal/analytics"
	"gitintel/backend/internal/cache"
	gh "gitintel/backend/internal/github"
	"gitintel/backend/internal/llm"
)

type barrierGitHubClient struct {
	started chan string
	release <-chan struct{}
}

func TestOptionalCheckLimitsReserveCoreGitHubQuota(t *testing.T) {
	tests := []struct {
		remaining int
		want      githubOptionalCheckLimits
	}{
		{remaining: 0, want: githubOptionalCheckLimits{}},
		{remaining: 5, want: githubOptionalCheckLimits{}},
		{remaining: 10, want: githubOptionalCheckLimits{}},
		{remaining: 13, want: githubOptionalCheckLimits{readmes: 3}},
		{remaining: 20, want: githubOptionalCheckLimits{readmes: 5, releases: 5}},
		{remaining: 30, want: githubOptionalCheckLimits{readmes: 5, releases: 5, testing: 5, contributors: 5}},
		{remaining: 35, want: githubOptionalCheckLimits{readmes: 5, releases: 5, testing: 5, contributors: 5, pullRequests: 5}},
		{remaining: 40, want: githubOptionalCheckLimits{readmes: 5, releases: 5, testing: 5, contributors: 5, pullRequests: 5, languages: 5}},
		{remaining: 45, want: githubOptionalCheckLimits{readmes: 5, releases: 5, testing: 5, contributors: 5, pullRequests: 5, languages: 10}},
	}
	for _, test := range tests {
		if got := optionalCheckLimits(test.remaining); got != test.want {
			t.Errorf("optionalCheckLimits(%d) = %+v, want %+v", test.remaining, got, test.want)
		}
	}
}

func (c *barrierGitHubClient) GetProfile(context.Context, string) (*gh.UserProfile, error) {
	c.started <- "profile"
	<-c.release
	return &gh.UserProfile{Username: "octocat"}, nil
}

func (c *barrierGitHubClient) GetRepositories(context.Context, string) ([]gh.Repository, error) {
	c.started <- "repositories"
	<-c.release
	return []gh.Repository{{Name: "project"}}, nil
}

func (c *barrierGitHubClient) RateLimit(context.Context) (gh.RateLimitInfo, error) {
	c.started <- "rate-limit"
	<-c.release
	return gh.RateLimitInfo{Remaining: 10}, nil
}

func TestFetchGitHubDataStartsIndependentRequestsConcurrently(t *testing.T) {
	release := make(chan struct{})
	client := &barrierGitHubClient{started: make(chan string, 3), release: release}
	type fetchResult struct {
		profile *gh.UserProfile
		repos   []gh.Repository
		limit   gh.RateLimitInfo
	}
	done := make(chan fetchResult, 1)
	go func() {
		profile, repos, limit, profileErr, reposErr := fetchGitHubData(context.Background(), client, "octocat", true)
		if profileErr != nil || reposErr != nil {
			done <- fetchResult{}
			return
		}
		done <- fetchResult{profile: profile, repos: repos, limit: limit}
	}()

	started := make(map[string]bool, 3)
	for range 3 {
		select {
		case request := <-client.started:
			started[request] = true
		case <-time.After(time.Second):
			close(release)
			t.Fatal("GitHub requests did not all start before any response was released")
		}
	}
	close(release)
	result := <-done
	if !started["profile"] || !started["repositories"] || !started["rate-limit"] {
		t.Fatalf("expected all three independent requests to start, got %v", started)
	}
	if result.profile == nil || len(result.repos) != 1 || result.limit.Remaining != 10 {
		t.Fatal("expected profile, repositories, and rate-limit data from the parallel fetch")
	}
}

func TestConcurrentAnalysisCacheMissesShareOneGitHubFetch(t *testing.T) {
	const callers = 12
	release := make(chan struct{})
	client := &barrierGitHubClient{started: make(chan string, callers*2), release: release}
	handler := NewHandler(cache.NewCache(AnalysisCacheTTL), analytics.NewEngine())
	var ready sync.WaitGroup
	ready.Add(callers)
	var completed sync.WaitGroup
	completed.Add(callers)
	errors := make(chan error, callers)
	for range callers {
		go func() {
			defer completed.Done()
			ready.Done()
			result, err := handler.loadOrFetchAnalysis(context.Background(), "octocat", client)
			if err != nil {
				errors <- err
				return
			}
			if result.Profile.Username != "octocat" || len(result.Repos) != 1 {
				errors <- fmt.Errorf("unexpected shared analysis result: %+v", result)
			}
		}()
	}
	ready.Wait()
	for range 2 {
		select {
		case <-client.started:
		case <-time.After(time.Second):
			close(release)
			t.Fatal("initial GitHub profile and repository fetch did not both start")
		}
	}
	time.Sleep(25 * time.Millisecond)
	select {
	case request := <-client.started:
		close(release)
		t.Fatalf("concurrent cache miss started duplicate GitHub request %q", request)
	default:
	}
	close(release)
	completed.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
	if len(client.started) != 0 {
		t.Fatalf("got %d unexpected GitHub fetch calls", len(client.started))
	}
}

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

func TestResumeAndPortfolioReuseCachedAnalysisWithoutGitHubRequests(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	client, err := gh.NewClientWithBaseURL(server.Client(), server.URL+"/", "")
	if err != nil {
		t.Fatalf("create mock GitHub client: %v", err)
	}

	store := cache.NewCache(AnalysisCacheTTL)
	store.Set(analysisCacheKey("octocat"), analyzeResult{
		Profile: gh.UserProfile{Username: "octocat"},
		Analysis: analytics.Analysis{
			ResumeMarkdown:        "cached resume",
			PortfolioMarkdown:     "cached portfolio",
			ProfileReadmeMarkdown: "cached profile readme",
		},
	})
	handler := NewHandler(store, analytics.NewEngine())
	handler.client = client

	for _, test := range []struct {
		path string
		want string
	}{
		{path: "/api/resume", want: "cached resume"},
		{path: "/api/portfolio", want: "cached portfolio"},
		{path: "/api/profile-readme", want: "cached profile readme"},
	} {
		request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(`{"username":"octocat"}`))
		response := httptest.NewRecorder()
		if test.path == "/api/resume" {
			handler.Resume(response, request)
		} else if test.path == "/api/portfolio" {
			handler.Portfolio(response, request)
		} else {
			handler.ProfileReadme(response, request)
		}
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), test.want) {
			t.Fatalf("%s returned status=%d body=%s", test.path, response.Code, response.Body.String())
		}
	}
	if requests != 0 {
		t.Fatalf("cached resume and portfolio requested GitHub %d times, want zero", requests)
	}
}

func TestResumeFetchesOnceAndPortfolioReusesNewAnalysis(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch strings.ToLower(r.URL.Path) {
		case "/users/octocat":
			_, _ = fmt.Fprint(w, `{"login":"octocat","public_repos":0}`)
		case "/users/octocat/repos":
			_, _ = fmt.Fprint(w, `[]`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := gh.NewClientWithBaseURL(server.Client(), server.URL+"/", "")
	if err != nil {
		t.Fatalf("create mock GitHub client: %v", err)
	}

	handler := NewHandler(cache.NewCache(AnalysisCacheTTL), analytics.NewEngine())
	handler.client = client
	resume := httptest.NewRecorder()
	handler.Resume(resume, httptest.NewRequest(http.MethodPost, "/api/resume", strings.NewReader(`{"username":"OctoCat"}`)))
	if resume.Code != http.StatusOK {
		t.Fatalf("resume returned status=%d body=%s", resume.Code, resume.Body.String())
	}
	portfolio := httptest.NewRecorder()
	handler.Portfolio(portfolio, httptest.NewRequest(http.MethodPost, "/api/portfolio", strings.NewReader(`{"username":"octocat"}`)))
	if portfolio.Code != http.StatusOK {
		t.Fatalf("portfolio returned status=%d body=%s", portfolio.Code, portfolio.Body.String())
	}
	if requests.Load() != 2 {
		t.Fatalf("resume followed by portfolio made %d GitHub calls, want exactly the two initial data requests", requests.Load())
	}
}

func TestInterpretRefetchesAnalysisAfterCacheExpiry(t *testing.T) {
	if AnalysisCacheTTL != 30*time.Minute {
		t.Fatalf("analysis cache TTL = %s, want 30m", AnalysisCacheTTL)
	}
	t.Setenv("LLM_ANALYSIS_ENABLED", "true")
	t.Setenv("LLM_PROVIDER_STATUS", "")
	t.Setenv("LLM_PROVIDER", "huggingface")
	t.Setenv("HF_API_KEYS", "cache-expiry-test-key")
	t.Setenv("HF_API_KEY", "")
	t.Setenv("HF_MODEL", "cache-expiry-test-model")

	var githubRequests atomic.Int32
	githubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		githubRequests.Add(1)
		switch strings.ToLower(r.URL.Path) {
		case "/users/octocat":
			_, _ = fmt.Fprint(w, `{"login":"octocat","public_repos":0}`)
		case "/users/octocat/repos":
			_, _ = fmt.Fprint(w, `[]`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer githubServer.Close()
	githubClient, err := gh.NewClientWithBaseURL(githubServer.Client(), githubServer.URL+"/", "")
	if err != nil {
		t.Fatalf("create mock GitHub client: %v", err)
	}

	llmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"choices":[{"message":{"content":"Fresh interpretation"}}]}`)
	}))
	defer llmServer.Close()
	t.Setenv("HF_BASE_URL", llmServer.URL)

	store := cache.NewCache(20 * time.Millisecond)
	store.Set(analysisCacheKey("octocat"), analyzeResult{
		Profile:  gh.UserProfile{Username: "octocat"},
		Evidence: llm.EvidenceContext{SchemaVersion: "1", Profile: llm.EvidenceProfile{Username: "stale"}},
	})
	time.Sleep(30 * time.Millisecond)
	credentials := llm.NewCredentialManager("HF_API_KEYS")
	credentials.Load()
	handler := NewHandlerWithCredentials(store, analytics.NewEngine(), credentials)
	handler.client = githubClient

	response := httptest.NewRecorder()
	handler.Interpret(response, httptest.NewRequest(http.MethodPost, "/api/interpret", strings.NewReader(`{"username":"octocat","language":"en"}`)))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Fresh interpretation") {
		t.Fatalf("interpret after expiry returned status=%d body=%s", response.Code, response.Body.String())
	}
	if githubRequests.Load() != 2 {
		t.Fatalf("expired analysis required %d GitHub requests, want profile and repositories only", githubRequests.Load())
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

func TestInterpretKnownDepletedModelMakesNoProviderRequest(t *testing.T) {
	t.Setenv("LLM_ANALYSIS_ENABLED", "true")
	t.Setenv("LLM_PROVIDER_STATUS", "")
	t.Setenv("LLM_PROVIDER", "huggingface")
	t.Setenv("HF_API_KEYS", "mock-test-key")
	t.Setenv("HF_API_KEY", "")
	t.Setenv("HF_MODEL", llm.LastKnownPaymentRequiredModel)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("HF_BASE_URL", server.URL)

	cacheStore := cache.NewCache(time.Minute)
	cacheStore.Set("analysis:octocat", analyzeResult{Profile: gh.UserProfile{Username: "octocat"}, Evidence: llm.EvidenceContext{SchemaVersion: "1"}})
	handler := NewHandler(cacheStore, analytics.NewEngine())
	response := httptest.NewRecorder()
	handler.Interpret(response, httptest.NewRequest(http.MethodPost, "/api/interpret", strings.NewReader(`{"username":"octocat","language":"en"}`)))
	if response.Code != http.StatusPaymentRequired || requests != 0 {
		t.Fatalf("known depleted model must be blocked without a provider request, status=%d requests=%d", response.Code, requests)
	}
}
