package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gh "github.com/google/go-github/v72/github"
)

func TestNormalizePublicRepositoryExcludesPrivateMetadata(t *testing.T) {
	privateFlag := true
	if _, include := normalizePublicRepository(&gh.Repository{Name: stringPointer("private-name"), Private: &privateFlag}); include {
		t.Fatal("private repository metadata must be excluded before analysis and evidence construction")
	}
	public, include := normalizePublicRepository(&gh.Repository{Name: stringPointer("public-name"), HTMLURL: stringPointer("https://github.com/octocat/public-name")})
	if !include || public.Name != "public-name" || public.URL == "" {
		t.Fatal("public repository metadata should be retained")
	}
	if _, include := normalizePublicRepository(nil); include {
		t.Fatal("nil repository entries must be excluded")
	}
}

func TestRepositoryListOptionsExcludeMemberRepositories(t *testing.T) {
	options := repositoryListOptions()
	if options.Type != "owner" {
		t.Fatalf("expected only owned repositories, got type %q", options.Type)
	}
	if options.PerPage != 100 {
		t.Fatalf("expected 100 repositories per page, got %d", options.PerPage)
	}
}

func TestCheckReadmePresenceWithMockServer(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/octocat/with-readme/readme", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/repos/octocat/no-readme/readme", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/repos/octocat/server-error/readme", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClientWithBaseURL(server.Client(), server.URL, "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	// 200 OK -> present
	present, err := client.CheckReadmePresence(ctx, "octocat", "with-readme")
	if err != nil {
		t.Fatalf("unexpected error for present readme: %v", err)
	}
	if !present {
		t.Fatal("expected present=true for HTTP 200")
	}

	// 404 Not Found -> absent
	absent, err := client.CheckReadmePresence(ctx, "octocat", "no-readme")
	if err != nil {
		t.Fatalf("unexpected error for absent readme: %v", err)
	}
	if absent {
		t.Fatal("expected present=false for HTTP 404")
	}

	// 500 Error -> error returned
	_, err = client.CheckReadmePresence(ctx, "octocat", "server-error")
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestVerifyCandidateReadmes(t *testing.T) {
	queried := make(map[string]int)
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/octocat/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		queried[path]++
		if strings.Contains(path, "repo-with-readme") {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClientWithBaseURL(server.Client(), server.URL, "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Create 7 repos; verify that only top 5 are queried
	now := time.Now()
	repos := []Repository{
		{Name: "repo-with-readme", FullName: "octocat/repo-with-readme", Stars: 100, UpdatedAt: now},
		{Name: "repo-2", FullName: "octocat/repo-2", Stars: 80, UpdatedAt: now},
		{Name: "repo-3", FullName: "octocat/repo-3", Stars: 60, UpdatedAt: now},
		{Name: "repo-4", FullName: "octocat/repo-4", Stars: 40, UpdatedAt: now},
		{Name: "repo-5", FullName: "octocat/repo-5", Stars: 20, UpdatedAt: now},
		{Name: "repo-6", FullName: "octocat/repo-6", Stars: 5, UpdatedAt: now},
		{Name: "repo-7", FullName: "octocat/repo-7", Stars: 1, UpdatedAt: now},
	}

	verified := client.VerifyCandidateReadmes(context.Background(), repos, 5)

	if len(queried) > 5 {
		t.Fatalf("expected at most 5 repos to be queried, got %d", len(queried))
	}

	// repo-with-readme should be verified present
	if verified[0].ReadmeStatus != ReadmeStatusVerifiedPresent || !verified[0].HasReadme {
		t.Fatalf("expected repo-with-readme to be verified_present and HasReadme=true, got status=%s, has_readme=%v",
			verified[0].ReadmeStatus, verified[0].HasReadme)
	}

	// repo-2 should be verified absent
	if verified[1].ReadmeStatus != ReadmeStatusVerifiedAbsent || verified[1].HasReadme {
		t.Fatalf("expected repo-2 to be verified_absent and HasReadme=false, got status=%s, has_readme=%v",
			verified[1].ReadmeStatus, verified[1].HasReadme)
	}

	// repo-6 and repo-7 should remain unverified
	if verified[5].ReadmeStatus != ReadmeStatusUnverified || verified[5].HasReadme {
		t.Fatalf("expected repo-6 to remain unverified, got status=%s", verified[5].ReadmeStatus)
	}
	if verified[6].ReadmeStatus != ReadmeStatusUnverified || verified[6].HasReadme {
		t.Fatalf("expected repo-7 to remain unverified, got status=%s", verified[6].ReadmeStatus)
	}
}

func TestRepositoryJSONMarshaling(t *testing.T) {
	repo := Repository{
		Name:         "test",
		HasReadme:    true,
		ReadmeStatus: ReadmeStatusVerifiedPresent,
	}
	bytes, err := json.Marshal(repo)
	if err != nil {
		t.Fatalf("failed to marshal repository: %v", err)
	}
	payload := string(bytes)
	if !strings.Contains(payload, `"has_readme":true`) {
		t.Fatalf("expected payload to contain \"has_readme\":true, got %s", payload)
	}
	if !strings.Contains(payload, `"readme_status":"verified_present"`) {
		t.Fatalf("expected payload to contain \"readme_status\":\"verified_present\", got %s", payload)
	}
}

func TestNormalizeRepositoryDoesNotHardcodeReadmeAbsent(t *testing.T) {
	// GI-DATA-006 Regression Test:
	// Verify that normalizeRepository initializes ReadmeStatus to ReadmeStatusUnverified,
	// rather than asserting negative evidence (verified absent).
	item := &gh.Repository{
		Name:     stringPointer("octocat-library"),
		FullName: stringPointer("octocat/octocat-library"),
	}
	repo := normalizeRepository(item)
	if repo.ReadmeStatus != ReadmeStatusUnverified {
		t.Fatalf("regression: expected normalizeRepository to mark README as %q, got %q",
			ReadmeStatusUnverified, repo.ReadmeStatus)
	}
	if repo.ReadmeStatus == ReadmeStatusVerifiedAbsent {
		t.Fatal("regression: normalizeRepository must not assert negative evidence (ReadmeStatusVerifiedAbsent) for uncollected README data")
	}

	// Verify that when candidate verification runs on normalized repositories with a present README,
	// HasReadme is updated to true and ReadmeStatus is updated to ReadmeStatusVerifiedPresent,
	// proving that the client does not permanently lock HasReadme to false.
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/octocat/octocat-library/readme", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client, err := NewClientWithBaseURL(server.Client(), server.URL, "")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	verified := client.VerifyCandidateReadmes(context.Background(), []Repository{repo}, 1)
	if len(verified) != 1 {
		t.Fatalf("expected 1 verified repository, got %d", len(verified))
	}
	if !verified[0].HasReadme {
		t.Fatal("regression: verified repository with README must have HasReadme=true, not hardcoded false")
	}
	if verified[0].ReadmeStatus != ReadmeStatusVerifiedPresent {
		t.Fatalf("regression: expected ReadmeStatus=%q, got %q",
			ReadmeStatusVerifiedPresent, verified[0].ReadmeStatus)
	}
}

func stringPointer(value string) *string { return &value }


