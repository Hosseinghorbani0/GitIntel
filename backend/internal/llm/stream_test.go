package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHuggingFaceProviderStreamsOpenAICompatibleDeltas(t *testing.T) {
	streamEnabled := make(chan bool, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Stream bool `json:"stream"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode provider request: %v", err)
		}
		streamEnabled <- request.Stream
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello \"}}]}\n\n")
		flusher.Flush()
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"world\"}}]}\n\n")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[],\"usage\":{\"total_tokens\":7}}\n\n")
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	provider := NewHuggingFaceProvider(Config{
		Provider:   "huggingface",
		Model:      "test-model",
		BaseURL:    server.URL,
		Credential: "test-key",
		Timeout:    5,
	})
	var tokens []string
	response, err := provider.StreamContext(context.Background(), Request{System: "system", User: "prompt"}, func(token string) error {
		tokens = append(tokens, token)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamContext returned error: %v", err)
	}
	if !<-streamEnabled {
		t.Fatal("expected upstream request to enable streaming")
	}
	if !reflect.DeepEqual(tokens, []string{"Hello ", "world"}) {
		t.Fatalf("streamed tokens = %q", tokens)
	}
	if response.Text != "Hello world" || response.TokensUsed != 7 {
		t.Fatalf("unexpected assembled response: %+v", response)
	}
}

func TestHuggingFaceProviderFallsBackToNonStreamingResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"choices":[{"message":{"content":"Complete response"}}],"usage":{"total_tokens":4}}`)
	}))
	defer server.Close()

	provider := NewHuggingFaceProvider(Config{
		Provider: "huggingface",
		Model:    "test-model",
		BaseURL:  server.URL,
		Timeout:  5,
	})
	var tokens []string
	response, err := provider.StreamContext(context.Background(), Request{System: "system", User: "prompt"}, func(token string) error {
		tokens = append(tokens, token)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamContext fallback returned error: %v", err)
	}
	if !reflect.DeepEqual(tokens, []string{"Complete response"}) {
		t.Fatalf("fallback callback received %q", tokens)
	}
	if response.Text != "Complete response" || response.TokensUsed != 4 {
		t.Fatalf("unexpected fallback response: %+v", response)
	}
}
