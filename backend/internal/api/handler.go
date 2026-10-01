package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"gitintel/backend/internal/analytics"
	"gitintel/backend/internal/cache"
	gh "gitintel/backend/internal/github"
	"gitintel/backend/internal/llm"
)

type Handler struct {
	cache  *cache.Cache
	engine *analytics.Engine
	client *gh.Client
}

func NewHandler(cacheStore *cache.Cache, engine *analytics.Engine) *Handler {
	return &Handler{cache: cacheStore, engine: engine, client: gh.NewClientFromToken("")}
}

type response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *apiError   `json:"error,omitempty"`
}

type apiError struct {
	Code       string                  `json:"code"`
	Message    string                  `json:"message"`
	Diagnostic *llm.ProviderDiagnostic `json:"diagnostic,omitempty"`
}

type analyzeRequest struct {
	Username string `json:"username"`
	Token    string `json:"token,omitempty"`
}

type analyzeResult struct {
	Profile   gh.UserProfile      `json:"profile"`
	Repos     []gh.Repository     `json:"repositories"`
	Analysis  analytics.Analysis  `json:"analysis"`
	Evidence  llm.EvidenceContext `json:"evidence"`
	RateLimit gh.RateLimitInfo    `json:"rate_limit"`
}

type githubDataClient interface {
	GetProfile(context.Context, string) (*gh.UserProfile, error)
	GetRepositories(context.Context, string) ([]gh.Repository, error)
	RateLimit(context.Context) (gh.RateLimitInfo, error)
}

func fetchGitHubData(ctx context.Context, client githubDataClient, username string, includeRateLimit bool) (*gh.UserProfile, []gh.Repository, gh.RateLimitInfo, error, error) {
	var profile *gh.UserProfile
	var repos []gh.Repository
	var rateLimit gh.RateLimitInfo
	var profileErr, reposErr error
	var wait sync.WaitGroup
	wait.Add(2)
	go func() {
		defer wait.Done()
		profile, profileErr = client.GetProfile(ctx, username)
	}()
	go func() {
		defer wait.Done()
		repos, reposErr = client.GetRepositories(ctx, username)
	}()
	if includeRateLimit {
		wait.Add(1)
		go func() {
			defer wait.Done()
			rateLimit, _ = client.RateLimit(ctx)
		}()
	}
	wait.Wait()
	return profile, repos, rateLimit, profileErr, reposErr
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, response{Success: true, Data: map[string]string{"status": "ok", "service": "GitIntel"}})
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/api/github/profile/")
	if username == "" {
		writeError(w, http.StatusBadRequest, "INVALID_USERNAME", "Username is required.")
		return
	}
	profile, err := h.loadProfile(username, "")
	if err != nil {
		writeError(w, githubErrorStatus(err), "GITHUB_PROFILE_ERROR", formatError(err))
		return
	}
	jsonResponse(w, http.StatusOK, response{Success: true, Data: profile})
}

func (h *Handler) Repositories(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/api/github/repos/")
	if username == "" {
		writeError(w, http.StatusBadRequest, "INVALID_USERNAME", "Username is required.")
		return
	}
	repos, err := h.loadRepos(username, "")
	if err != nil {
		writeError(w, githubErrorStatus(err), "GITHUB_REPO_ERROR", formatError(err))
		return
	}
	jsonResponse(w, http.StatusOK, response{Success: true, Data: repos})
}

func (h *Handler) Analyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST is required.")
		return
	}
	var req analyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Request body is invalid.")
		return
	}
	username := strings.TrimSpace(req.Username)
	if username == "" {
		writeError(w, http.StatusBadRequest, "INVALID_USERNAME", "GitHub username is required.")
		return
	}
	client := gh.NewClientFromToken(strings.TrimSpace(req.Token))
	profile, repos, limit, profileErr, reposErr := fetchGitHubData(r.Context(), client, username, true)
	if profileErr != nil {
		writeError(w, githubErrorStatus(profileErr), "GITHUB_PROFILE_ERROR", formatError(profileErr))
		return
	}
	if reposErr != nil {
		writeError(w, githubErrorStatus(reposErr), "GITHUB_REPO_ERROR", formatError(reposErr))
		return
	}
	analysis := h.engine.Analyze(*profile, repos)
	evidence := llm.BuildLLMContext(*profile, repos, analysis)
	result := analyzeResult{Profile: *profile, Repos: repos, Analysis: analysis, Evidence: evidence, RateLimit: limit}
	cacheKey := "analysis:" + username
	h.cache.Set(cacheKey, result)
	jsonResponse(w, http.StatusOK, response{Success: true, Data: result})
}

func (h *Handler) Interpret(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST is required.")
		return
	}
	var req struct {
		Username string `json:"username"`
		Language string `json:"language"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Request body is invalid.")
		return
	}
	username := strings.TrimSpace(req.Username)
	language := strings.ToLower(strings.TrimSpace(req.Language))
	if username == "" {
		writeError(w, http.StatusBadRequest, "INVALID_USERNAME", "GitHub username is required.")
		return
	}
	if language != "en" && language != "fa" {
		writeError(w, http.StatusBadRequest, "INVALID_LANGUAGE", "Language must be en or fa.")
		return
	}
	var result analyzeResult
	if !h.cache.Get("analysis:"+username, &result) {
		writeError(w, http.StatusNotFound, "ANALYSIS_NOT_FOUND", "Run deterministic profile analysis before requesting an interpretation.")
		return
	}
	if !llm.AnalysisReportsEnabled() {
		writeError(w, http.StatusServiceUnavailable, "LLM_ANALYSIS_DISABLED", "Deterministic analysis is available. AI interpretation is disabled until a provider is explicitly enabled.")
		return
	}
	cfg, err := llm.LoadConfigFromEnv()
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_LLM_CONFIG", "The LLM provider configuration is invalid.")
		return
	}
	interpreted, err := llm.GenerateInterpretation(r.Context(), cfg, result.Evidence, language)
	if err != nil {
		writeInterpretationError(w, err)
		return
	}
	jsonResponse(w, http.StatusOK, response{Success: true, Data: map[string]interface{}{
		"status":     "success",
		"language":   language,
		"provider":   interpreted.Provider,
		"model":      interpreted.Model,
		"report":     interpreted.Text,
		"latency_ms": interpreted.LatencyMS,
		"tokens":     interpreted.TokensUsed,
	}})
}

func writeInterpretationError(w http.ResponseWriter, err error) {
	status := http.StatusBadGateway
	code := "LLM_PROVIDER_ERROR"
	var providerError *llm.ProviderError
	switch {
	case errors.Is(err, llm.ErrAnalysisDisabled):
		status, code = http.StatusServiceUnavailable, "LLM_ANALYSIS_DISABLED"
	case llm.IsInvalidCredential(err):
		status, code = http.StatusUnauthorized, "LLM_INVALID_CREDENTIAL"
	case llm.IsPaymentRequired(err):
		status, code = http.StatusPaymentRequired, "LLM_PAYMENT_REQUIRED"
	case llm.IsRateLimited(err):
		status, code = http.StatusTooManyRequests, "LLM_RATE_LIMITED"
	case errors.As(err, &providerError) && providerError.Category == llm.ErrorUnauthorized:
		status, code = http.StatusForbidden, "LLM_FORBIDDEN"
	case errors.As(err, &providerError) && providerError.Category == llm.ErrorTimeout:
		status, code = http.StatusGatewayTimeout, "LLM_TIMEOUT"
	}
	message := "AI interpretation is unavailable. Deterministic analysis remains available."
	if providerError != nil && providerError.Diagnostic != nil {
		message = providerError.Error()
	}
	jsonResponse(w, status, response{Success: false, Error: &apiError{Code: code, Message: message, Diagnostic: providerErrorDiagnostic(providerError)}})
}

func providerErrorDiagnostic(err *llm.ProviderError) *llm.ProviderDiagnostic {
	if err == nil {
		return nil
	}
	return err.Diagnostic
}

func (h *Handler) Report(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/api/report/")
	if username == "" {
		writeError(w, http.StatusBadRequest, "INVALID_USERNAME", "Username is required.")
		return
	}
	cacheKey := "analysis:" + username
	var result analyzeResult
	if !h.cache.Get(cacheKey, &result) {
		writeError(w, http.StatusNotFound, "REPORT_NOT_FOUND", "No cached report available for this username.")
		return
	}
	jsonResponse(w, http.StatusOK, response{Success: true, Data: result.Analysis.ReportMarkdown})
}

func (h *Handler) Resume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST is required.")
		return
	}
	var req analyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Request body is invalid.")
		return
	}
	client := gh.NewClientFromToken(strings.TrimSpace(req.Token))
	profile, repos, _, profileErr, reposErr := fetchGitHubData(r.Context(), client, req.Username, false)
	if profileErr != nil {
		writeError(w, githubErrorStatus(profileErr), "GITHUB_PROFILE_ERROR", formatError(profileErr))
		return
	}
	if reposErr != nil {
		writeError(w, githubErrorStatus(reposErr), "GITHUB_REPO_ERROR", formatError(reposErr))
		return
	}
	analysis := h.engine.Analyze(*profile, repos)
	jsonResponse(w, http.StatusOK, response{Success: true, Data: map[string]string{"text": analysis.ResumeMarkdown}})
}

func (h *Handler) Portfolio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST is required.")
		return
	}
	var req analyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Request body is invalid.")
		return
	}
	client := gh.NewClientFromToken(strings.TrimSpace(req.Token))
	profile, repos, _, profileErr, reposErr := fetchGitHubData(r.Context(), client, req.Username, false)
	if profileErr != nil {
		writeError(w, githubErrorStatus(profileErr), "GITHUB_PROFILE_ERROR", formatError(profileErr))
		return
	}
	if reposErr != nil {
		writeError(w, githubErrorStatus(reposErr), "GITHUB_REPO_ERROR", formatError(reposErr))
		return
	}
	analysis := h.engine.Analyze(*profile, repos)
	jsonResponse(w, http.StatusOK, response{Success: true, Data: map[string]string{"text": analysis.PortfolioMarkdown}})
}

func (h *Handler) loadProfile(username, token string) (*gh.UserProfile, error) {
	client := gh.NewClientFromToken(token)
	ctx, cancel := contextWithTimeout(15 * time.Second)
	defer cancel()
	return client.GetProfile(ctx, username)
}

func (h *Handler) loadRepos(username, token string) ([]gh.Repository, error) {
	client := gh.NewClientFromToken(token)
	ctx, cancel := contextWithTimeout(15 * time.Second)
	defer cancel()
	return client.GetRepositories(ctx, username)
}

func contextWithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

func formatError(err error) string {
	if err == nil {
		return "GitHub request failed."
	}
	if strings.Contains(err.Error(), "rate limit") || strings.Contains(err.Error(), "429") {
		return "GitHub API rate limit reached. Try again later or use a token with a higher quota."
	}
	if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
		return "GitHub user was not found or is not publicly accessible."
	}
	if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "Bad credentials") {
		return "The provided GitHub token is invalid or lacks access to the requested data."
	}
	return "GitHub data could not be loaded. Please verify the username and network access."
}

func githubErrorStatus(err error) int {
	if err == nil {
		return http.StatusBadGateway
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "rate limit") || strings.Contains(message, "429") {
		return http.StatusTooManyRequests
	}
	if strings.Contains(message, "404") || strings.Contains(message, "not found") {
		return http.StatusNotFound
	}
	if strings.Contains(message, "401") || strings.Contains(message, "bad credentials") {
		return http.StatusUnauthorized
	}
	if strings.Contains(message, "403") {
		return http.StatusForbidden
	}
	return http.StatusBadGateway
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response{Success: false, Error: &apiError{Code: code, Message: message}})
}

func jsonResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

var ErrNotFound = errors.New("not found")
