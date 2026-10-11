package llm

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"gitintel/backend/internal/analytics"
	gh "gitintel/backend/internal/github"
)

func TestBuildLLMContextIsBoundedAndContainsObservedSignals(t *testing.T) {
	profile := gh.UserProfile{Username: "public-user", PublicRepos: 20, Followers: 4}
	repos := make([]gh.Repository, 15)
	for index := range repos {
		repos[index] = gh.Repository{
			Name:        "repo-" + string(rune('a'+index)),
			Description: "must not be sent to interpretation",
			Language:    "Go",
			Stars:       index,
			UpdatedAt:   time.Now(),
		}
	}
	analysis := analytics.NewEngine().Analyze(profile, repos)
	evidence := BuildLLMContext(profile, repos, analysis)
	if !evidence.RepositoryDataUntrusted || len(evidence.Repositories.Sample) != evidenceRepositorySampleLimit {
		t.Fatal("expected bounded repository sample marked as untrusted")
	}
	if len(evidence.Signals) != 6 || len(evidence.Signals[0].Metric) == 0 || len(evidence.Signals[0].Limitations) == 0 {
		t.Fatal("expected all analyzer signals with metrics and limitations")
	}
	if len(evidence.Signals[0].Repositories) == 0 {
		t.Fatal("expected signal-to-repository references in the evidence schema")
	}
	encoded, err := json.Marshal(evidence)
	if err != nil {
		t.Fatalf("marshal evidence: %v", err)
	}
	if strings.Contains(string(encoded), "must not be sent to interpretation") {
		t.Fatal("repository descriptions must not be included in LLM evidence")
	}
	if !strings.Contains(string(encoded), "public-user") || !strings.Contains(string(encoded), "Testing Evidence") {
		t.Fatal("expected actual profile identity and deterministic signals in evidence")
	}
}

func TestEvidenceSchemaV2BoundsAndPrioritizesSampleWithCollectionStates(t *testing.T) {
	profile := gh.UserProfile{Username: "public-user", PublicRepos: 500}
	repos := make([]gh.Repository, 500)
	for index := range repos {
		repos[index] = gh.Repository{
			Name:          fmt.Sprintf("repo-%d", index),
			FullName:      fmt.Sprintf("public-user/repo-%d", index),
			Stars:         index,
			UpdatedAt:     time.Now(),
			ReadmeStatus:  gh.ReadmeStatusUnverified,
			ReleaseStatus: gh.ReleaseStatusUnverified,
			TestingStatus: gh.TestingArtifactsUnverified,
			Fork:           index%10 == 0,
			Archived:       index%5 == 0,
		}
	}
	repos[499].ReadmeStatus = gh.ReadmeStatusVerifiedPresent
	repos[499].HasReadme = true
	repos[499].ReadmeCheckAttempted = true
	repos[499].ReleaseStatus = gh.ReleaseStatusVerifiedPresent
	repos[499].HasReleases = true
	repos[499].ReleaseCheckAttempted = true
	repos[499].TestingStatus = gh.TestingArtifactsVerifiedPresent
	repos[499].TestingArtifacts = []string{"tests"}
	repos[499].TestingCheckAttempted = true
	analysis := analytics.NewEngine().Analyze(profile, repos)
	evidence := BuildLLMContext(profile, repos, analysis)

	if evidence.SchemaVersion != "2" {
		t.Fatalf("expected evidence schema version 2, got %q", evidence.SchemaVersion)
	}
	if evidence.Repositories.Total != 500 || len(evidence.Repositories.Sample) != evidenceRepositorySampleLimit {
		t.Fatalf("expected full aggregate and capped sample, got total=%d sample=%d", evidence.Repositories.Total, len(evidence.Repositories.Sample))
	}
	if evidence.Repositories.Original != 450 || evidence.Repositories.Forked != 50 || evidence.Repositories.Archived != 100 {
		t.Fatalf("expected accurate full-list aggregates, got original=%d forked=%d archived=%d", evidence.Repositories.Original, evidence.Repositories.Forked, evidence.Repositories.Archived)
	}
	if evidence.Repositories.Sample[0].FullName != "public-user/repo-499" {
		t.Fatalf("expected highest-scoring repository first, got %q", evidence.Repositories.Sample[0].FullName)
	}
	if !evidence.Repositories.ReadmeVerificationPerformed || !evidence.Repositories.ReleasesRead || !evidence.Repositories.TestingArtifactsInspected {
		t.Fatal("expected queried collection states to be represented in evidence")
	}
	if evidence.CollectionParameters.RepositorySampleMaximum != evidenceRepositorySampleLimit ||
		evidence.CollectionParameters.RepositoryPagesMaximum != 3 {
		t.Fatalf("expected explicit collection parameters, got %#v", evidence.CollectionParameters)
	}
	encoded, err := json.Marshal(evidence)
	if err != nil {
		t.Fatalf("marshal evidence: %v", err)
	}
	for _, field := range []string{
		`"readme_verification_performed":true`,
		`"release_endpoints_queried":true`,
		`"testing_artifacts_inspected":true`,
		`"repository_pages_maximum":3`,
		`"readme_repositories_attempted":1`,
		`"release_repositories_attempted":1`,
		`"testing_repositories_attempted":1`,
	} {
		if !strings.Contains(string(encoded), field) {
			t.Fatalf("expected schema field %s in serialized evidence", field)
		}
	}
}

func TestEvidenceIncludesExplicitFalseForUnqueriedCollections(t *testing.T) {
	profile := gh.UserProfile{Username: "public-user"}
	repos := []gh.Repository{{Name: "repo", FullName: "public-user/repo"}}
	evidence := BuildLLMContext(profile, repos, analytics.NewEngine().Analyze(profile, repos))
	encoded, err := json.Marshal(evidence)
	if err != nil {
		t.Fatalf("marshal evidence: %v", err)
	}
	for _, field := range []string{
		`"readme_verification_performed":false`,
		`"release_endpoints_queried":false`,
		`"testing_artifacts_inspected":false`,
		`"readme_repositories_attempted":0`,
		`"release_repositories_attempted":0`,
		`"testing_repositories_attempted":0`,
	} {
		if !strings.Contains(string(encoded), field) {
			t.Fatalf("expected explicit unqueried state %s", field)
		}
	}
}

func TestReportPromptUsesSuppliedEvidenceAndLanguageOnlyVaries(t *testing.T) {
	evidence := map[string]interface{}{
		"profile":   map[string]interface{}{"username": "actual-public-user"},
		"signals":   []string{"Maintenance: Moderate"},
		"untrusted": true,
	}
	english := ReportPrompt(evidence, "en")
	persian := ReportPrompt(evidence, "fa")
	if english.Language != "en" || persian.Language != "fa" {
		t.Fatal("report prompt language must match requested language")
	}
	if english.Input.(map[string]interface{})["profile"].(map[string]interface{})["username"] != "actual-public-user" ||
		persian.Input.(map[string]interface{})["profile"].(map[string]interface{})["username"] != "actual-public-user" {
		t.Fatal("both languages must use the caller-supplied analysis evidence")
	}
	if strings.Contains(english.User, "example-project") || strings.Contains(persian.User, "example-project") {
		t.Fatal("report prompt must not use the model-test fixture")
	}
	if !strings.Contains(persian.System, "natural Persian") {
		t.Fatal("Persian report prompt must request natural professional Persian")
	}
}

func TestBuildProviderSupportsGenericOpenAICompatibleConfig(t *testing.T) {
	provider, err := BuildProvider(Config{Provider: "openai-compatible", BaseURL: "https://provider.example/v1", Model: "model-x"})
	if err != nil {
		t.Fatalf("build generic provider: %v", err)
	}
	if provider.Name() != "openai-compatible" {
		t.Fatalf("expected configured provider name, got %q", provider.Name())
	}
	if _, err := BuildProvider(Config{Provider: "unknown-vendor", BaseURL: "https://provider.example/v1", Model: "model-x"}); err == nil {
		t.Fatal("unsupported provider must not silently route through Hugging Face")
	}
}
