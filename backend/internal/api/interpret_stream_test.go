package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"gitintel/backend/internal/analytics"
	"gitintel/backend/internal/cache"
	gh "gitintel/backend/internal/github"
	"gitintel/backend/internal/llm"
)

func TestInterpretStreamReturnsProgressAndCompletionEvents(t *testing.T) {
	t.Setenv("LLM_ANALYSIS_ENABLED", "true")
	t.Setenv("LLM_PROVIDER_STATUS", "")
	t.Setenv("LLM_PROVIDER", "huggingface")
	t.Setenv("HF_API_KEYS", "stream-test-key")
	t.Setenv("HF_API_KEY", "")
	t.Setenv("HF_MODEL", "stream-test-model")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Stream bool `json:"stream"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode upstream request: %v", err)
		}
		if !request.Stream {
			t.Error("expected upstream streaming to be enabled")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"First \"}}]}\n\n")
		w.(http.Flusher).Flush()
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"token\"}}]}\n\n")
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer upstream.Close()
	t.Setenv("HF_BASE_URL", upstream.URL)

	store := cache.NewCache(time.Minute)
	store.Set("analysis:octocat", analyzeResult{
		Profile:  gh.UserProfile{Username: "octocat"},
		Evidence: llm.EvidenceContext{SchemaVersion: "1", Profile: llm.EvidenceProfile{Username: "octocat"}},
	})
	credentials := llm.NewCredentialManager("HF_API_KEYS")
	credentials.Load()
	handler := NewHandlerWithCredentials(store, analytics.NewEngine(), credentials)
	request := httptest.NewRequest(http.MethodGet, "/api/interpret/stream?username=octocat&language=en", nil)
	response := httptest.NewRecorder()

	handler.InterpretStream(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatalf("unexpected response headers: status=%d headers=%v", response.Code, response.Header())
	}
	body := response.Body.String()
	for _, expected := range []string{
		"event: token\ndata: {\"text\":\"First \"}",
		"event: token\ndata: {\"text\":\"token\"}",
		"event: done\ndata: {\"language\":\"en\"",
		"\"report\":\"First token\"",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("stream body is missing %q: %s", expected, body)
		}
	}
}

func TestInterpretStreamFlushesFirstTokenWithin500Milliseconds(t *testing.T) {
	t.Setenv("LLM_ANALYSIS_ENABLED", "true")
	t.Setenv("LLM_PROVIDER_STATUS", "")
	t.Setenv("LLM_PROVIDER", "huggingface")
	t.Setenv("HF_API_KEYS", "stream-test-key")
	t.Setenv("HF_API_KEY", "")
	t.Setenv("HF_MODEL", "stream-test-model")
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseUpstream := func() { releaseOnce.Do(func() { close(release) }) }
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"first\"}}]}\n\n")
		w.(http.Flusher).Flush()
		<-release
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\" second\"}}]}\n\n")
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	store := cache.NewCache(time.Minute)
	store.Set("analysis:octocat", analyzeResult{
		Profile:  gh.UserProfile{Username: "octocat"},
		Evidence: llm.EvidenceContext{SchemaVersion: "1", Profile: llm.EvidenceProfile{Username: "octocat"}},
	})
	credentials := llm.NewCredentialManager("HF_API_KEYS")
	credentials.Load()
	handler := NewHandlerWithCredentials(store, analytics.NewEngine(), credentials)
	server := httptest.NewServer(http.HandlerFunc(handler.InterpretStream))
	var response *http.Response
	defer func() {
		releaseUpstream()
		if response != nil {
			_ = response.Body.Close()
		}
		server.Close()
		upstream.Close()
	}()
	t.Setenv("HF_BASE_URL", upstream.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/interpret/stream?username=octocat&language=en", nil)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	response, err = server.Client().Do(request)
	if err != nil {
		t.Fatalf("request stream: %v", err)
	}
	reader := bufio.NewReader(response.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read first streamed token: %v", err)
		}
		if strings.HasPrefix(line, "event: token") {
			break
		}
	}
	firstTokenLatency := time.Since(started)
	if firstTokenLatency >= 500*time.Millisecond {
		t.Fatalf("first SSE token arrived after %s, expected under 500ms", firstTokenLatency)
	}
	releaseUpstream()
	remaining, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read remaining stream: %v", err)
	}
	if !strings.Contains(string(remaining), "event: done") {
		t.Fatalf("stream did not complete after the first token: %s", remaining)
	}
}
