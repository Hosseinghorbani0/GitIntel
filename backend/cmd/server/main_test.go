package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithCORSAllowsOnlyLocalFrontendOrigins(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := withCORS(next)

	for _, origin := range []string{"http://localhost:5173", "http://127.0.0.1:5173"} {
		t.Run(origin, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			request.Header.Set("Origin", origin)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if got := response.Header().Get("Access-Control-Allow-Origin"); got != origin {
				t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, origin)
			}
			if got := response.Code; got != http.StatusOK {
				t.Fatalf("status = %d, want %d", got, http.StatusOK)
			}
		})
	}
}

func TestWithCORSDoesNotAllowUnlistedOrigin(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	request := httptest.NewRequest(http.MethodOptions, "/api/health", nil)
	request.Header.Set("Origin", "https://attacker.example")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unlisted origin received Access-Control-Allow-Origin %q", got)
	}
	if got := response.Code; got != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", got, http.StatusNoContent)
	}
}
