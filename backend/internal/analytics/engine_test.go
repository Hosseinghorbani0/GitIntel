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

func TestBuildLanguageDistributionUsesLanguageByteVolume(t *testing.T) {
	repos := []gh.Repository{
		{Name: "alpha", Language: "Go", FullName: "owner/alpha", LanguageBytes: map[string]int{"Go": 900, "TypeScript": 100}},
		{Name: "beta", Language: "Python", FullName: "owner/beta", LanguageBytes: map[string]int{"Python": 2000}},
	}

	langs := buildLanguageDistribution(repos)
	if len(langs) != 3 {
		t.Fatalf("expected 3 language buckets, got %d", len(langs))
	}
	byName := map[string]LanguageBucket{}
	for _, lang := range langs {
		byName[lang.Name] = lang
	}
	if byName["Python"].Count != 2000 {
		t.Fatalf("expected Python byte volume to be 2000, got %d", byName["Python"].Count)
	}
	if byName["Go"].Count != 900 {
		t.Fatalf("expected Go byte volume to be 900, got %d", byName["Go"].Count)
	}
	if byName["TypeScript"].Count != 100 {
		t.Fatalf("expected TypeScript byte volume to be 100, got %d", byName["TypeScript"].Count)
	}
	if byName["Python"].Percentage <= byName["Go"].Percentage {
		t.Fatalf("expected Python to dominate when weighted by bytes, got Go=%f Python=%f", byName["Go"].Percentage, byName["Python"].Percentage)
	}
}

func TestBuildLanguageDistributionUsesOnlySuccessfulByteSample(t *testing.T) {
	repos := []gh.Repository{
		{Name: "sampled", Language: "Go", LanguageBytes: map[string]int{"Go": 100}},
		{Name: "not-sampled", Language: "Python"},
	}

	langs := buildLanguageDistribution(repos)
	if len(langs) != 1 || langs[0].Name != "Go" || langs[0].Percentage != 100 {
		t.Fatalf("expected byte distribution to exclude unmeasured repository-count data, got %#v", langs)
	}
	basis, sampleSize := languageDistributionSource(repos)
	if basis != "code_bytes_sampled" || sampleSize != 1 {
		t.Fatalf("expected sampled byte source with one measured repository, got %q and %d", basis, sampleSize)
	}
}

func TestLanguageDistributionSourceUsesRepositoryFrequencyWhenUnmeasured(t *testing.T) {
	repos := []gh.Repository{
		{Name: "alpha", Language: "Go"},
		{Name: "beta", Language: "Python"},
	}

	basis, sampleSize := languageDistributionSource(repos)
	if basis != "repository_frequency_fallback" || sampleSize != 0 {
		t.Fatalf("expected repository-frequency fallback, got %q and %d", basis, sampleSize)
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
	if analysis.ResumeMarkdown == "" || analysis.PortfolioMarkdown == "" || analysis.ReportMarkdown == "" || analysis.ProfileReadmeMarkdown == "" || analysis.AIPromptMarkdown == "" {
		t.Fatal("resume, portfolio, report, profile readme, and AI prompt markdown should all be generated")
	}
	if !strings.Contains(analysis.ProfileReadmeMarkdown, "Engineering Signals") || !strings.Contains(analysis.ProfileReadmeMarkdown, "GitIntel") {
		t.Fatal("profile readme should contain verified signals and GitIntel attribution")
	}
	if !strings.Contains(analysis.AIPromptMarkdown, "VERIFIED ENGINEERING SIGNALS") || !strings.Contains(analysis.AIPromptMarkdown, "INSTRUCTION") {
		t.Fatal("AI prompt should contain structured verification sections and instructions")
	}
	if !strings.Contains(analysis.ReportMarkdown, "Observed metric") || !strings.Contains(analysis.ReportMarkdown, "Limitation:") || !strings.Contains(analysis.ReportMarkdown, "AI authorship") {
		t.Fatal("deterministic report must retain metrics and explicit limitations")
	}
	if len(analysis.FeaturedProjects) == 0 {
		t.Fatal("expected at least one featured repository")
	}
}

func TestTestingEvidenceUsesVerifiedArtifactsNotRepositoryNames(t *testing.T) {
	sandbox := computeTestingEvidence([]gh.Repository{{
		Name: "test-sandbox", TestingStatus: gh.TestingArtifactsVerifiedAbsent,
	}})
	if sandbox.Level != "Limited" || sandbox.Value != 0 || sandbox.Metric["repositories_with_test_artifacts"] != 0 {
		t.Fatalf("repository name alone must not create a testing signal, got %#v", sandbox)
	}
	verified := computeTestingEvidence([]gh.Repository{{
		Name: "service", FullName: "octocat/service",
		TestingStatus:    gh.TestingArtifactsVerifiedPresent,
		TestingArtifacts: []string{"internal/service_test.go"},
	}})
	if verified.Level != "Moderate" || len(verified.Repositories) != 1 || verified.Repositories[0] != "octocat/service" {
		t.Fatalf("verified artifacts must be cited with their repository, got %#v", verified)
	}
	if !containsString(verified.Limitations, "Observed test directory structures and CI configuration; test coverage percentages and assertion rigor were not evaluated.") {
		t.Fatalf("expected explicit testing quality limitations, got %v", verified.Limitations)
	}
	if computeTestingEvidence([]gh.Repository{{Name: "unknown"}}).Level != "Insufficient data" {
		t.Fatal("uncollected test artifacts must remain insufficient data")
	}
}

func TestComputeDocumentationRejectsNameAndDescriptionHeuristics(t *testing.T) {
	signal := computeDocumentation([]gh.Repository{{
		Name: "docs-broken", Description: "documentation docs",
	}})
	if signal.Level != "Limited" || signal.Metric["repositories_with_documentation_artifacts"] != 0 {
		t.Fatalf("unverified name or description hints must not earn documentation credit, got %#v", signal)
	}
	withFlags := computeDocumentation([]gh.Repository{{
		Name: "project", HasPages: true, HasWiki: true, HasLicense: true,
	}})
	if withFlags.Metric["repositories_with_github_pages"] != 1 ||
		withFlags.Metric["repositories_with_github_wiki"] != 1 ||
		withFlags.Metric["repositories_with_license"] != 1 {
		t.Fatalf("expected verified documentation artifacts to be counted, got %#v", withFlags.Metric)
	}
}

func TestCollaborationDoesNotCountPersonalForksAsEvidence(t *testing.T) {
	idleFork := computeCollaboration([]gh.Repository{{Name: "my-fork", Fork: true}})
	if idleFork.Level != "Limited" || idleFork.Metric["personal_forked_repositories"] != 1 {
		t.Fatalf("a personal fork without external activity must remain Limited, got %#v", idleFork)
	}
	original := computeCollaboration([]gh.Repository{{Name: "shared", FullName: "octocat/shared", Forks: 3}})
	if original.Level != "Moderate" || original.Metric["incoming_forks_on_original_repositories"] != 3 {
		t.Fatalf("incoming forks on original repositories should be distinguished, got %#v", original)
	}
	if _, exists := original.Metric["open_issues"]; exists {
		t.Fatal("combined issue and pull-request count must not use the ambiguous open_issues key")
	}
}

func TestCollaborationUsesVerifiedContributorCounts(t *testing.T) {
	singleContributor := computeCollaboration([]gh.Repository{{
		Name: "solo", FullName: "octocat/solo", ContributorsCount: 1,
		ContributorsStatus: gh.ContributorsVerified,
	}})
	if singleContributor.Level != "Limited" || singleContributor.Metric["repositories_with_verified_contributor_counts"] != 1 {
		t.Fatalf("a single verified contributor should not imply multi-person collaboration, got %#v", singleContributor)
	}

	multiContributor := computeCollaboration([]gh.Repository{{
		Name: "team-project", FullName: "octocat/team-project", ContributorsCount: 4,
		ContributorsStatus: gh.ContributorsVerified,
	}})
	if multiContributor.Level != "Moderate" ||
		multiContributor.Metric["repositories_with_multiple_contributors"] != 1 ||
		multiContributor.Metric["contributors_observed_on_bounded_sample"] != 4 {
		t.Fatalf("multiple verified contributors should be recorded as collaboration evidence, got %#v", multiContributor)
	}
	if len(multiContributor.Repositories) != 1 || multiContributor.Repositories[0] != "octocat/team-project" {
		t.Fatalf("expected contributor evidence to reference its repository, got %v", multiContributor.Repositories)
	}
}

func TestCollaborationExposesSeparatedIssueAndPullRequestCounts(t *testing.T) {
	signal := computeCollaboration([]gh.Repository{{
		Name: "active", FullName: "octocat/active", OpenIssues: 5,
		OpenIssuesCount: 3, OpenPullRequests: 2,
		PullRequestCountStatus: gh.PullRequestCountsVerified,
	}})
	if signal.Metric["open_issues_and_prs"] != 5 ||
		signal.Metric["verified_open_issues_on_bounded_sample"] != 3 ||
		signal.Metric["verified_open_pull_requests_on_bounded_sample"] != 2 ||
		signal.Metric["repositories_with_verified_issue_pr_counts"] != 1 {
		t.Fatalf("expected separate verified issue/PR counts while retaining combined total, got %#v", signal.Metric)
	}
}

func TestArchivedRepositoriesAreNotClassifiedAsNeglected(t *testing.T) {
	signal := computeMaintenance([]gh.Repository{{Name: "stable-library", Archived: true}})
	if signal.Classification != "Archived / Completed" || !strings.Contains(signal.Evidence[0], "inactivity alone is not treated as poor maintenance") {
		t.Fatalf("archived repositories must be treated as potentially complete, got %#v", signal)
	}
}

func TestCodeEvolutionDisclosesTimestampProxies(t *testing.T) {
	signal := computeCodeEvolution([]gh.Repository{{Name: "active", FullName: "octocat/active", UpdatedAt: time.Now()}})
	if !containsString(signal.Limitations, "Tracks recent push and repository update timestamps. Commit volume, diff size, and branch activity are not collected in this REST collection.") {
		t.Fatalf("expected timestamp proxy limitation, got %v", signal.Limitations)
	}
	if len(signal.Repositories) != 1 || signal.Repositories[0] != "octocat/active" {
		t.Fatalf("expected supporting repository reference, got %v", signal.Repositories)
	}
}

func TestComputeDocumentationWithTriStateReadme(t *testing.T) {
	repos := []gh.Repository{
		{
			Name:         "repo-verified-readme",
			ReadmeStatus: gh.ReadmeStatusVerifiedPresent,
			HasReadme:    true,
		},
		{
			Name:         "repo-verified-absent",
			ReadmeStatus: gh.ReadmeStatusVerifiedAbsent,
			HasReadme:    false,
		},
		{
			Name:         "repo-unverified-1",
			ReadmeStatus: gh.ReadmeStatusUnverified,
			HasReadme:    false,
		},
		{
			Name:         "repo-unverified-2",
			ReadmeStatus: gh.ReadmeStatusUnverified,
			HasReadme:    false,
		},
	}

	signal := computeDocumentation(repos)
	if signal.Level != "Moderate" {
		t.Fatalf("expected Moderate documentation signal, got %s", signal.Level)
	}

	if signal.Metric["repositories_with_verified_readme"] != 1 {
		t.Fatalf("expected 1 repo with verified readme, got %f", signal.Metric["repositories_with_verified_readme"])
	}
	if signal.Metric["repositories_with_unverified_readme"] != 2 {
		t.Fatalf("expected 2 repos with unverified readme, got %f", signal.Metric["repositories_with_unverified_readme"])
	}

	foundLimitation := false
	for _, lim := range signal.Limitations {
		if strings.Contains(lim, "unverified for 2 repositories") {
			foundLimitation = true
			break
		}
	}
	if !foundLimitation {
		t.Fatalf("expected limitations to disclose unverified count, got %v", signal.Limitations)
	}

	foundEvidence := false
	for _, ev := range signal.Evidence {
		if strings.Contains(ev, "1 repositories have verified README") {
			foundEvidence = true
			break
		}
	}
	if !foundEvidence {
		t.Fatalf("expected evidence to cite verified README, got %v", signal.Evidence)
	}
}

func TestSelectFeaturedBoostsVerifiedReadme(t *testing.T) {
	now := time.Now()
	repos := []gh.Repository{
		{
			Name:         "documented",
			FullName:     "owner/documented",
			Stars:        10,
			UpdatedAt:    now,
			ReadmeStatus: gh.ReadmeStatusVerifiedPresent,
			HasReadme:    true,
		},
		{
			Name:         "undocumented",
			FullName:     "owner/undocumented",
			Stars:        10,
			UpdatedAt:    now,
			ReadmeStatus: gh.ReadmeStatusVerifiedAbsent,
			HasReadme:    false,
		},
	}

	featured := selectFeatured(repos)
	if len(featured) != 2 {
		t.Fatalf("expected 2 featured repos, got %d", len(featured))
	}
	if featured[0].Repository.Name != "documented" {
		t.Fatalf("expected documented repo to rank first due to verified README boost; got %s", featured[0].Repository.Name)
	}

	foundWhy := false
	for _, why := range featured[0].Why {
		if why == "Verified README" {
			foundWhy = true
			break
		}
	}
	if !foundWhy {
		t.Fatalf("expected 'Verified README' in why explanation, got %v", featured[0].Why)
	}
}

func TestComputeReleaseActivityWithTriStateReleases(t *testing.T) {
	// 1. Two verified present releases -> Strong
	reposStrong := []gh.Repository{
		{Name: "core-api", ReleaseStatus: gh.ReleaseStatusVerifiedPresent, HasReleases: true, ReleaseCount: 5},
		{Name: "web-client", ReleaseStatus: gh.ReleaseStatusVerifiedPresent, HasReleases: true, ReleaseCount: 2},
		{Name: "helper-lib", ReleaseStatus: gh.ReleaseStatusUnverified},
	}
	signalStrong := computeReleaseActivity(reposStrong)
	if signalStrong.Level != "Strong" {
		t.Fatalf("expected Strong level for 2 verified releases, got %s", signalStrong.Level)
	}
	if signalStrong.Metric["repositories_with_verified_releases"] != 2 {
		t.Fatalf("expected 2 verified releases in metric, got %f", signalStrong.Metric["repositories_with_verified_releases"])
	}
	if signalStrong.Metric["repositories_with_unverified_release"] != 1 {
		t.Fatalf("expected 1 unverified release in metric, got %f", signalStrong.Metric["repositories_with_unverified_release"])
	}
	if len(signalStrong.Evidence) == 0 || !strings.Contains(signalStrong.Evidence[0], "core-api, web-client") {
		t.Fatalf("expected evidence to cite repository names, got %v", signalStrong.Evidence)
	}

	// 2. One verified present release -> Moderate
	reposModerate := []gh.Repository{
		{Name: "core-api", ReleaseStatus: gh.ReleaseStatusVerifiedPresent, HasReleases: true, ReleaseCount: 1},
		{Name: "docs-site", ReleaseStatus: gh.ReleaseStatusVerifiedAbsent},
	}
	signalModerate := computeReleaseActivity(reposModerate)
	if signalModerate.Level != "Moderate" {
		t.Fatalf("expected Moderate level for 1 verified release, got %s", signalModerate.Level)
	}

	// 3. All verified absent -> Limited (honest negative evidence)
	reposAbsent := []gh.Repository{
		{Name: "script-1", ReleaseStatus: gh.ReleaseStatusVerifiedAbsent},
		{Name: "script-2", ReleaseStatus: gh.ReleaseStatusVerifiedAbsent},
	}
	signalAbsent := computeReleaseActivity(reposAbsent)
	if signalAbsent.Level != "Limited" {
		t.Fatalf("expected Limited level for verified absent releases, got %s", signalAbsent.Level)
	}

	// 4. Unverified repos -> Insufficient data (never penalized)
	reposUnverified := []gh.Repository{
		{Name: "repo-1", ReleaseStatus: gh.ReleaseStatusUnverified},
		{Name: "repo-2", ReleaseStatus: gh.ReleaseStatusUnverified},
	}
	signalUnverified := computeReleaseActivity(reposUnverified)
	if signalUnverified.Level != "Insufficient data" {
		t.Fatalf("expected Insufficient data for unverified releases, got %s", signalUnverified.Level)
	}
	foundLimitation := false
	for _, lim := range signalUnverified.Limitations {
		if strings.Contains(lim, "Release status was unverified for 2 repositories") {
			foundLimitation = true
			break
		}
	}
	if !foundLimitation {
		t.Fatalf("expected limitations to disclose unverified count, got %v", signalUnverified.Limitations)
	}
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func TestSelectFeaturedBoostsVerifiedReleases(t *testing.T) {
	now := time.Now()
	repos := []gh.Repository{
		{
			Name:          "with-releases",
			FullName:      "owner/with-releases",
			Stars:         10,
			UpdatedAt:     now,
			ReleaseStatus: gh.ReleaseStatusVerifiedPresent,
			HasReleases:   true,
			ReleaseCount:  3,
		},
		{
			Name:          "without-releases",
			FullName:      "owner/without-releases",
			Stars:         10,
			UpdatedAt:     now,
			ReleaseStatus: gh.ReleaseStatusVerifiedAbsent,
			HasReleases:   false,
		},
	}

	featured := selectFeatured(repos)
	if len(featured) != 2 {
		t.Fatalf("expected 2 featured repos, got %d", len(featured))
	}
	if featured[0].Repository.Name != "with-releases" {
		t.Fatalf("expected with-releases to rank first due to release boost; got %s", featured[0].Repository.Name)
	}

	foundWhy := false
	for _, why := range featured[0].Why {
		if why == "Published releases" {
			foundWhy = true
			break
		}
	}
	if !foundWhy {
		t.Fatalf("expected 'Published releases' in why explanation, got %v", featured[0].Why)
	}
}
