package analytics

import (
	"fmt"
	"strings"
	"testing"
	"time"

	gh "gitintel/backend/internal/github"
)

func BenchmarkAnalyze1000Repositories(b *testing.B) {
	now := time.Now()
	languages := []string{"Go", "TypeScript", "Python", "Rust"}
	repos := make([]gh.Repository, 1000)
	for i := range repos {
		repos[i] = gh.Repository{
			Name:       fmt.Sprintf("repo-%d", i),
			FullName:   fmt.Sprintf("octocat/repo-%d", i),
			Language:   languages[i%len(languages)],
			Stars:      i % 100,
			Forks:      i % 20,
			UpdatedAt:  now.Add(-time.Duration(i%365) * 24 * time.Hour),
			PushedAt:   now.Add(-time.Duration(i%365) * 24 * time.Hour),
			Fork:       i%5 == 0,
			Archived:   i%17 == 0,
			HasReadme:  i%2 == 0,
			HasLicense: i%3 == 0,
		}
	}
	engine := NewEngine()
	profile := gh.UserProfile{Username: "octocat", DisplayName: "Octocat"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Analyze(profile, repos)
	}
}

func TestBuildLanguageDistributionUsesRepositoryCounts(t *testing.T) {
	repos := []gh.Repository{
		{Name: "alpha", Language: "Go", FullName: "owner/alpha"},
		{Name: "beta", Language: "Go", FullName: "owner/beta"},
		{Name: "gamma", Language: "TypeScript", FullName: "owner/gamma"},
	}

	langs := buildLanguageDistribution(repos)
	if len(langs) != 2 {
		t.Fatalf("expected 2 language buckets, got %d", len(langs))
	}
	if langs[0].Name != "Go" {
		t.Fatalf("expected Go to lead the distribution, got %s", langs[0].Name)
	}
	if langs[0].Percentage <= 50 || langs[0].Percentage > 100 {
		t.Fatalf("unexpected percentage for Go: %f", langs[0].Percentage)
	}
}

func TestSelectFeaturedPrefersActiveDocumentedRepos(t *testing.T) {
	now := time.Now()
	repos := []gh.Repository{
		{Name: "quiet", FullName: "owner/quiet", Stars: 2, Forks: 0, UpdatedAt: now.AddDate(0, -1, 0), HasReadme: true, HasLicense: true},
		{Name: "active", FullName: "owner/active", Stars: 44, Forks: 6, UpdatedAt: now.AddDate(0, 0, -10), HasReadme: true, HasLicense: true},
		{Name: "archived", FullName: "owner/archived", Stars: 5, Forks: 1, UpdatedAt: now.AddDate(0, -7, 0), HasReadme: true, Archived: true},
	}

	featured := selectFeatured(repos)
	if len(featured) != 3 {
		t.Fatalf("expected 3 featured projects, got %d", len(featured))
	}
	if featured[0].Repository.Name != "active" {
		t.Fatalf("expected the most active repo to rank first; got %s", featured[0].Repository.Name)
	}
}

func TestSelectFeaturedKeepsOnlyHighestScoringRepositories(t *testing.T) {
	repos := make([]gh.Repository, 10)
	for i := range repos {
		repos[i] = gh.Repository{Name: fmt.Sprintf("project-%d", i), Stars: i}
	}

	featured := selectFeatured(repos)
	if len(featured) != 5 {
		t.Fatalf("expected five featured projects, got %d", len(featured))
	}
	for i, project := range featured {
		want := fmt.Sprintf("project-%d", 9-i)
		if project.Repository.Name != want {
			t.Fatalf("featured project %d = %q, want %q", i, project.Repository.Name, want)
		}
	}
}

func TestAnalyzeGeneratesSignalsAndCopyableOutputs(t *testing.T) {
	profile := gh.UserProfile{Username: "octocat", DisplayName: "Octocat", ProfileURL: "https://github.com/octocat", Followers: 42, Following: 7}
	repos := []gh.Repository{
		{Name: "webapp", FullName: "octocat/webapp", Description: "App", Language: "TypeScript", Stars: 42, Forks: 5, UpdatedAt: time.Now().AddDate(0, 0, -15), HasReadme: true, HasLicense: true, HasDocs: true},
		{Name: "api", FullName: "octocat/api", Description: "API service", Language: "Go", Stars: 20, Forks: 1, UpdatedAt: time.Now().AddDate(0, 0, -30), HasReadme: true, HasLicense: true, HasDocs: true},
		{Name: "tests", FullName: "octocat/tests", Description: "Test suite", Language: "Python", Stars: 5, Forks: 1, UpdatedAt: time.Now().AddDate(0, 0, -10), HasReadme: true, HasLicense: true},
	}

	analysis := NewEngine().Analyze(profile, repos)
	if analysis.Summary == "" {
		t.Fatal("summary should not be empty")
	}
	if len(analysis.Signals) != 6 {
		t.Fatalf("expected six signal categories, got %d", len(analysis.Signals))
	}
	for _, signal := range analysis.Signals {
		if signal.Classification == "" || len(signal.Metric) == 0 || len(signal.Evidence) == 0 || len(signal.Limitations) == 0 {
			t.Errorf("signal %q must retain classification, observed metric, evidence, and limitations", signal.Name)
		}
	}
	if analysis.ResumeMarkdown == "" || analysis.PortfolioMarkdown == "" || analysis.ReportMarkdown == "" {
		t.Fatal("resume, portfolio, and report markdown should all be generated")
	}
	if !strings.Contains(analysis.ReportMarkdown, "Observed metric") || !strings.Contains(analysis.ReportMarkdown, "Limitation:") || !strings.Contains(analysis.ReportMarkdown, "AI authorship") {
		t.Fatal("deterministic report must retain metrics and explicit limitations")
	}
	if len(analysis.FeaturedProjects) == 0 {
		t.Fatal("expected at least one featured repository")
	}
}

func TestTestingEvidenceDisclosesRepositoryNameHeuristic(t *testing.T) {
	withoutMarker := computeTestingEvidence([]gh.Repository{{Name: "service"}})
	withMarker := computeTestingEvidence([]gh.Repository{{Name: "service-tests"}})
	if !strings.Contains(withoutMarker.Evidence[0], "contents and CI workflows were not inspected") {
		t.Fatal("missing test-name marker should disclose uninspected repository contents")
	}
	if !strings.Contains(withMarker.Evidence[0], "name heuristics") || !strings.Contains(withMarker.Evidence[0], "test files and CI workflows were not inspected") {
		t.Fatal("test-name marker should be reported as a heuristic, not verified test artifacts")
	}
}
