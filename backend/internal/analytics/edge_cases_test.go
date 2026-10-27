
package analytics

// GI-TEST-002: Edge-case test suite for abnormal developer profiles.
// Validates that the analytics engine does not panic, divide-by-zero,
// or produce deceptive signals when given boundary-condition inputs.
//
// Edge cases tested:
//   1. Zero public repositories
//   2. 100% forked repositories (zero original work)
//   3. 100% archived repositories
//   4. Repositories with zero stars, forks, and open issues
//   5. Non-English repository names and descriptions (Unicode / RTL / mixed)
//   6. Extremely large repository count (1,000+ repos)

import (
	"strings"
	"testing"
	"time"

	gh "gitintel/backend/internal/github"
)

// --- Edge Case 1: Zero public repositories ---

func TestAnalyzeZeroRepositoriesDoesNotPanic(t *testing.T) {
	profile := gh.UserProfile{Username: "empty-user", DisplayName: "Empty User"}
	result := NewEngine().Analyze(profile, nil)

	if result.Summary == "" {
		t.Fatal("summary must not be empty even for a profile with zero repositories")
	}
	if len(result.Signals) != 6 {
		t.Fatalf("expected 6 signals even for zero repositories, got %d", len(result.Signals))
	}
	for _, signal := range result.Signals {
		if signal.Classification == "" {
			t.Errorf("signal %q must have a classification even for zero repositories", signal.Name)
		}
	}
	if len(result.FeaturedProjects) != 0 {
		t.Errorf("expected no featured projects for zero repositories, got %d", len(result.FeaturedProjects))
	}
}

func TestAnalyzeEmptyRepositorySliceDoesNotPanic(t *testing.T) {
	profile := gh.UserProfile{Username: "empty-user", DisplayName: "Empty User"}
	result := NewEngine().Analyze(profile, []gh.Repository{})

	if result.Summary == "" {
		t.Fatal("summary must not be empty for empty repository slice")
	}
	if result.ResumeMarkdown == "" || result.PortfolioMarkdown == "" {
		t.Fatal("resume and portfolio markdown must be generated even for empty repository list")
	}
}

func TestLanguageDistributionForZeroRepositoriesIsEmpty(t *testing.T) {
	langs := buildLanguageDistribution(nil)
	if len(langs) != 0 {
		t.Fatalf("expected empty language distribution for nil repos, got %d buckets", len(langs))
	}
	langs = buildLanguageDistribution([]gh.Repository{})
	if len(langs) != 0 {
		t.Fatalf("expected empty language distribution for empty repos, got %d buckets", len(langs))
	}
}

// --- Edge Case 2: 100% forked repositories ---

func TestAnalyzeAllForksDoesNotExaggerateCollaboration(t *testing.T) {
	now := time.Now()
	repos := []gh.Repository{
		{Name: "fork-1", FullName: "user/fork-1", Fork: true, Stars: 0, UpdatedAt: now},
		{Name: "fork-2", FullName: "user/fork-2", Fork: true, Stars: 0, UpdatedAt: now},
		{Name: "fork-3", FullName: "user/fork-3", Fork: true, Stars: 5, UpdatedAt: now},
	}
	profile := gh.UserProfile{Username: "fork-collector", DisplayName: "Fork Collector"}
	result := NewEngine().Analyze(profile, repos)

	var collaborationSignal *Signal
	for i, s := range result.Signals {
		if s.Name == "Collaboration" {
			collaborationSignal = &result.Signals[i]
			break
		}
	}
	if collaborationSignal == nil {
		t.Fatal("expected a Collaboration signal")
	}
	// All repos are personal forks — no genuine collaboration signal
	if collaborationSignal.Level == "Strong" {
		t.Errorf("all-fork profile must not receive Strong collaboration; got %q", collaborationSignal.Level)
	}
	if collaborationSignal.Metric["personal_forked_repositories"] != float64(len(repos)) {
		t.Errorf("expected all repos counted as personal forks, got metric=%v", collaborationSignal.Metric)
	}
}

func TestSelectFeaturedWithAllForksIncludesThemIfNothingElseExists(t *testing.T) {
	now := time.Now()
	repos := []gh.Repository{
		{Name: "fork-a", FullName: "user/fork-a", Fork: true, Stars: 10, UpdatedAt: now},
		{Name: "fork-b", FullName: "user/fork-b", Fork: true, Stars: 5, UpdatedAt: now},
	}
	featured := selectFeatured(repos)
	if len(featured) == 0 {
		t.Fatal("selectFeatured must not return empty results when all repos are forks — they are still a developer's portfolio")
	}
}

// --- Edge Case 3: 100% archived repositories ---

func TestAnalyzeAllArchivedDoesNotClassifyAsNeglected(t *testing.T) {
	repos := []gh.Repository{
		{Name: "old-lib", FullName: "user/old-lib", Archived: true, Stars: 50, UpdatedAt: time.Now().AddDate(-3, 0, 0)},
		{Name: "legacy-api", FullName: "user/legacy-api", Archived: true, Stars: 10, UpdatedAt: time.Now().AddDate(-5, 0, 0)},
	}
	profile := gh.UserProfile{Username: "archivist", DisplayName: "The Archivist"}
	result := NewEngine().Analyze(profile, repos)

	var maintenanceSignal *Signal
	for i, s := range result.Signals {
		if s.Name == "Maintenance" {
			maintenanceSignal = &result.Signals[i]
			break
		}
	}
	if maintenanceSignal == nil {
		t.Fatal("expected a Code Maintenance signal")
	}
	if maintenanceSignal.Classification == "Neglected" {
		t.Errorf("all-archived profile must not be classified as 'Neglected'; got %q", maintenanceSignal.Classification)
	}
	// Must be "Archived / Completed"
	if !strings.Contains(maintenanceSignal.Classification, "Archived") {
		t.Errorf("all-archived profile classification must reference Archived status; got %q", maintenanceSignal.Classification)
	}
}

func TestComputeMaintenanceAllArchivedUsesArchivedClassification(t *testing.T) {
	repos := []gh.Repository{
		{Name: "stable-v1", Archived: true},
		{Name: "stable-v2", Archived: true},
		{Name: "stable-v3", Archived: true},
	}
	signal := computeMaintenance(repos)
	if signal.Classification != "Archived / Completed" {
		t.Errorf("expected 'Archived / Completed' classification, got %q", signal.Classification)
	}
	found := false
	for _, ev := range signal.Evidence {
		if strings.Contains(ev, "inactivity alone is not treated as poor maintenance") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected evidence to clarify inactivity vs. poor maintenance, got %v", signal.Evidence)
	}
}

// --- Edge Case 4: Repositories with zero stars, forks, open issues ---

func TestAnalyzeZeroMetricRepositoriesDoesNotDivideByZero(t *testing.T) {
	now := time.Now()
	repos := []gh.Repository{
		{Name: "zero-repo-1", FullName: "user/zero-repo-1", Stars: 0, Forks: 0, OpenIssues: 0, UpdatedAt: now},
		{Name: "zero-repo-2", FullName: "user/zero-repo-2", Stars: 0, Forks: 0, OpenIssues: 0, UpdatedAt: now},
	}
	profile := gh.UserProfile{Username: "beginner", DisplayName: "Beginner Dev"}

	// Must not panic
	result := NewEngine().Analyze(profile, repos)
	if result.Summary == "" {
		t.Fatal("summary must not be empty for zero-metric repositories")
	}
	for _, signal := range result.Signals {
		if signal.Classification == "" {
			t.Errorf("signal %q must have a classification even for zero-metric repos", signal.Name)
		}
	}
}

func TestSelectFeaturedWithZeroStarRepositoriesDoesNotPanic(t *testing.T) {
	repos := []gh.Repository{
		{Name: "project", FullName: "user/project", Stars: 0, Forks: 0},
	}
	// Must not panic and must return the one repo
	featured := selectFeatured(repos)
	if len(featured) != 1 {
		t.Fatalf("expected 1 featured project for 1 zero-star repo, got %d", len(featured))
	}
}

func TestCollaborationWithZeroOpenIssuesAndZeroForks(t *testing.T) {
	repos := []gh.Repository{
		{Name: "solo-project", FullName: "user/solo-project", OpenIssues: 0, Forks: 0, Fork: false},
	}
	signal := computeCollaboration(repos)
	if signal.Classification == "" {
		t.Fatal("collaboration signal must always have a classification")
	}
	// Zero engagement should be Limited, not panicking
	if signal.Level != "Limited" {
		t.Errorf("zero-engagement project expected Limited collaboration, got %q", signal.Level)
	}
}

// --- Edge Case 5: Non-English names and descriptions (Unicode / RTL / mixed) ---

func TestNormalizeRepositoryWithUnicodeNamesDoesNotCorruptData(t *testing.T) {
	unicodeRepos := []gh.Repository{
		// Persian / RTL repo name and description
		{
			Name: "پروژه-آزمایشی", FullName: "user/پروژه-آزمایشی",
			Description: "یک پروژه آزمایشی برای بررسی کدها",
			Language:    "Python",
			Stars:       5,
			UpdatedAt:   time.Now(),
		},
		// Arabic-script mixed with Latin
		{
			Name: "arabic-مشروع", FullName: "user/arabic-مشروع",
			Description: "مشروع اختبار — Test Project",
			Language:    "Go",
			Stars:       10,
			UpdatedAt:   time.Now(),
		},
		// Japanese
		{
			Name: "プロジェクト", FullName: "user/プロジェクト",
			Description: "テスト用のプロジェクトです",
			Language:    "TypeScript",
			Stars:       3,
			UpdatedAt:   time.Now(),
		},
		// Emoji in description
		{
			Name: "cool-project", FullName: "user/cool-project",
			Description: "🚀 A blazing fast ⚡ tool with 💯 coverage",
			Language:    "Rust",
			Stars:       42,
			UpdatedAt:   time.Now(),
		},
	}

	profile := gh.UserProfile{Username: "polyglot-dev", DisplayName: "Polyglot Developer"}
	result := NewEngine().Analyze(profile, unicodeRepos)

	if result.Summary == "" {
		t.Fatal("summary must not be empty for Unicode-named repositories")
	}
	if len(result.Signals) != 6 {
		t.Fatalf("expected 6 signals for Unicode repos, got %d", len(result.Signals))
	}
	for _, signal := range result.Signals {
		if signal.Classification == "" {
			t.Errorf("signal %q must have a classification for Unicode repos", signal.Name)
		}
	}
	if result.ResumeMarkdown == "" {
		t.Fatal("resume markdown must be generated for Unicode-named repos without corruption")
	}
}

func TestSelectFeaturedWithUnicodeNamesDoesNotPanic(t *testing.T) {
	now := time.Now()
	repos := []gh.Repository{
		{Name: "پروژه-آزمایشی", FullName: "user/پروژه-آزمایشی", Stars: 50, UpdatedAt: now},
		{Name: "プロジェクト", FullName: "user/プロジェクト", Stars: 30, UpdatedAt: now},
	}
	// Must not panic; Unicode repo names must survive selectFeatured
	featured := selectFeatured(repos)
	if len(featured) == 0 {
		t.Fatal("selectFeatured must return results for Unicode-named repos")
	}
	if featured[0].Repository.Name != "پروژه-آزمایشی" {
		t.Errorf("expected higher-star repo to rank first; got %q", featured[0].Repository.Name)
	}
}

func TestLanguageDistributionWithEmptyLanguageFieldDoesNotPanic(t *testing.T) {
	repos := []gh.Repository{
		{Name: "no-lang-1", Language: ""},
		{Name: "no-lang-2", Language: ""},
		{Name: "go-project", Language: "Go"},
	}
	langs := buildLanguageDistribution(repos)
	// Repos with empty Language must not generate a bucket with name ""
	for _, bucket := range langs {
		if bucket.Name == "" {
			t.Error("language distribution must not include an empty-string language bucket")
		}
	}
}

// --- Edge Case 6: Extremely large repository count (1,000+ repos) ---

func TestAnalyze1000RepositoriesDoesNotPanic(t *testing.T) {
	now := time.Now()
	languages := []string{"Go", "TypeScript", "Python", "Rust", "Java", "Swift", ""}
	repos := make([]gh.Repository, 1000)
	for i := range repos {
		repos[i] = gh.Repository{
			Name:        strings.Repeat("x", i%50+1),
			FullName:    "megauser/" + strings.Repeat("x", i%50+1),
			Description: "Repository number " + string(rune('0'+i%10)),
			Language:    languages[i%len(languages)],
			Stars:       i % 500,
			Forks:       i % 100,
			OpenIssues:  i % 50,
			UpdatedAt:   now.Add(-time.Duration(i) * time.Hour),
			PushedAt:    now.Add(-time.Duration(i) * time.Hour),
			Fork:        i%7 == 0,
			Archived:    i%23 == 0,
		}
	}

	profile := gh.UserProfile{
		Username:    "megauser",
		DisplayName: "Mega User",
		PublicRepos: 1000,
	}

	// Must complete without panicking or divide-by-zero
	result := NewEngine().Analyze(profile, repos)

	if result.Summary == "" {
		t.Fatal("summary must not be empty for 1000-repo profile")
	}
	if len(result.Signals) != 6 {
		t.Fatalf("expected 6 signals for 1000 repos, got %d", len(result.Signals))
	}
	// Featured projects must be capped at 5
	if len(result.FeaturedProjects) > 5 {
		t.Errorf("featured projects must be capped at 5, got %d", len(result.FeaturedProjects))
	}
}

func TestLanguageDistributionWith1000RepositoriesDoesNotExceed100Percent(t *testing.T) {
	languages := []string{"Go", "TypeScript", "Python", "Rust", "Java", "C++", "Ruby"}
	repos := make([]gh.Repository, 1000)
	for i := range repos {
		repos[i] = gh.Repository{
			Name:     "repo",
			Language: languages[i%len(languages)],
		}
	}
	langs := buildLanguageDistribution(repos)

	total := 0.0
	for _, bucket := range langs {
		if bucket.Percentage < 0 {
			t.Errorf("language %q has negative percentage %f", bucket.Name, bucket.Percentage)
		}
		total += bucket.Percentage
	}
	// Allow small floating-point variance
	if total < 99.9 || total > 100.1 {
		t.Errorf("language distribution percentages must sum to ~100%%, got %.2f%%", total)
	}
}

func TestSelectFeaturedWith1000RepositoriesReturnsCappedResults(t *testing.T) {
	now := time.Now()
	repos := make([]gh.Repository, 1000)
	for i := range repos {
		repos[i] = gh.Repository{
			Name:      strings.Repeat("r", i%10+1),
			FullName:  "user/" + strings.Repeat("r", i%10+1),
			Stars:     i,
			UpdatedAt: now.Add(-time.Duration(i) * time.Hour),
		}
	}
	featured := selectFeatured(repos)
	if len(featured) > 5 {
		t.Errorf("selectFeatured must return at most 5 projects, got %d", len(featured))
	}
	if len(featured) == 0 {
		t.Fatal("selectFeatured must return at least 1 project from 1000 repos")
	}
	// Results must be in descending score order (highest stars last since repos[999] has Stars=999)
	if featured[0].Repository.Stars < featured[len(featured)-1].Repository.Stars {
		t.Errorf("expected featured projects in descending score order")
	}
}
