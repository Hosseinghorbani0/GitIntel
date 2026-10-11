package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"gitintel/backend/internal/llm"
)

func (h *Handler) InterpretStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET is required.")
		return
	}
	username := strings.TrimSpace(r.URL.Query().Get("username"))
	language := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("language")))
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
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "STREAMING_UNAVAILABLE", "The server does not support streaming responses.")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, ": connected\n\n"); err != nil {
		return
	}
	flusher.Flush()

	interpreted, err := llm.StreamInterpretationWithManager(r.Context(), cfg, result.Evidence, language, h.credentials, func(token string) error {
		_, err := writeSSEEvent(w, flusher, "token", map[string]string{"text": token})
		return err
	})
	if err != nil {
		if r.Context().Err() == nil {
			_, _ = writeSSEEvent(w, flusher, "error", map[string]string{"code": interpretationErrorCode(err)})
		}
		return
	}
	if interpreted == nil {
		_, _ = writeSSEEvent(w, flusher, "error", map[string]string{"code": "LLM_PROVIDER_ERROR"})
		return
	}
	_, _ = writeSSEEvent(w, flusher, "done", map[string]interface{}{
		"status":     "success",
		"language":   language,
		"provider":   interpreted.Provider,
		"model":      interpreted.Model,
		"report":     interpreted.Text,
		"latency_ms": interpreted.LatencyMS,
		"tokens":     interpreted.TokensUsed,
	})
}

func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, event string, payload interface{}) (int, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	written, err := io.WriteString(w, "event: "+event+"\ndata: "+string(encoded)+"\n\n")
	if err != nil {
		return written, err
	}
	flusher.Flush()
	return written, nil
}

func interpretationErrorCode(err error) string {
	var providerError *llm.ProviderError
	switch {
	case errors.Is(err, llm.ErrAnalysisDisabled):
		return "LLM_ANALYSIS_DISABLED"
	case llm.IsInvalidCredential(err):
		return "LLM_INVALID_CREDENTIAL"
	case llm.IsPaymentRequired(err):
		return "LLM_PAYMENT_REQUIRED"
	case llm.IsRateLimited(err):
		return "LLM_RATE_LIMITED"
	case errors.As(err, &providerError) && providerError.Category == llm.ErrorUnauthorized:
		return "LLM_FORBIDDEN"
	case errors.As(err, &providerError) && providerError.Category == llm.ErrorTimeout:
		return "LLM_TIMEOUT"
	default:
		return "LLM_PROVIDER_ERROR"
	}
}
