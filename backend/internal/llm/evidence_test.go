package llm

import (
	"encoding/json"
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
