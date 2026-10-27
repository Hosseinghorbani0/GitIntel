package github

// GI-TEST-001: Fixture-based integration tests asserting that the real GitHub client
// normalization pipeline processes raw API JSON correctly, including the fields that
// were historically hardcoded or never populated (HasReadme, HasReleases).
//
// The JSON fixtures in testdata/ are sampled from real GitHub REST API responses
// and are fed through normalizePublicRepository + normalizeRepository without
// any hand-massaging, ensuring that production client behavior is covered.

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	gh "github.com/google/go-github/v72/github"
)

// loadRepositoryFixture reads a JSON fixture from testdata/ and unmarshals it
// into the go-github library's raw []*gh.Repository slice, mirroring what the
// GitHub REST client receives from the wire before normalization.
func loadRepositoryFixture(t *testing.T, filename string) []*gh.Repository {
	t.Helper()
	data, err := os.ReadFile("testdata/" + filename)
	if err != nil {
		t.Fatalf("load fixture %q: %v", filename, err)
	}
	var repos []*gh.Repository
	if err := json.Unmarshal(data, &repos); err != nil {
		t.Fatalf("unmarshal fixture %q: %v", filename, err)
	}
	return repos
}

// loadProfileFixture reads a JSON fixture from testdata/ and unmarshals it into
// the go-github library's raw *gh.User, mirroring the wire response for /users/{username}.
func loadProfileFixture(t *testing.T, filename string) *gh.User {
	t.Helper()
	data, err := os.ReadFile("testdata/" + filename)
	if err != nil {
		t.Fatalf("load fixture %q: %v", filename, err)
	}
	var user gh.User
	if err := json.Unmarshal(data, &user); err != nil {
		t.Fatalf("unmarshal fixture %q: %v", filename, err)
	}
	return &user
}

// TestNormalizeRepositoryFixtureDoesNotHardcodeReadmeAbsent (GI-TEST-001 regression):
// Feeds the real_octocat_repos.json fixture through normalizePublicRepository and
// asserts that no repository is assigned ReadmeStatusVerifiedAbsent or HasReadme=true
// from the raw API response alone — because the GitHub repo list endpoint never
// returns a has_readme field, so all values must be ReadmeStatusUnverified.
func TestNormalizeRepositoryFixtureDoesNotHardcodeReadmeAbsent(t *testing.T) {
	rawRepos := loadRepositoryFixture(t, "real_octocat_repos.json")
	if len(rawRepos) == 0 {
		t.Fatal("fixture must contain at least one repository")
	}

	for _, item := range rawRepos {
		repo, include := normalizePublicRepository(item)
		if item.GetPrivate() {
			if include {
				t.Errorf("private repository %q must not be included in normalized output", item.GetName())
			}
			continue
		}
		if !include {
			t.Errorf("public repository %q was unexpectedly excluded by normalization", item.GetName())
			continue
		}

		// The canonical regression assertion from GI-DATA-006:
		// normalizeRepository must never assert ReadmeStatusVerifiedAbsent for
		// data that was never collected from the GitHub API.
		if repo.ReadmeStatus == ReadmeStatusVerifiedAbsent {
			t.Errorf("repo %q: normalizeRepository must not assert negative README evidence (ReadmeStatusVerifiedAbsent) for uncollected data", repo.Name)
		}
		if repo.ReadmeStatus != ReadmeStatusUnverified {
			t.Errorf("repo %q: expected ReadmeStatus=%q from raw API fixture, got %q", repo.Name, ReadmeStatusUnverified, repo.ReadmeStatus)
		}
		if repo.HasReadme {
			t.Errorf("repo %q: HasReadme must be false before bounded verification runs", repo.Name)
		}
	}
}

// TestNormalizeRepositoryFixtureDoesNotHardcodeReleaseAbsent (GI-TEST-001 regression):
// Similar to the README check, no repository in the raw fixture should have
// ReleaseStatusVerifiedAbsent — releases are only marked absent after
// VerifyCandidateReleases explicitly queries the /releases endpoint.
func TestNormalizeRepositoryFixtureDoesNotHardcodeReleaseAbsent(t *testing.T) {
	rawRepos := loadRepositoryFixture(t, "real_octocat_repos.json")

	for _, item := range rawRepos {
		if item.GetPrivate() {
			continue
		}
		repo, include := normalizePublicRepository(item)
		if !include {
			continue
		}
		if repo.ReleaseStatus == ReleaseStatusVerifiedAbsent {
			t.Errorf("repo %q: normalizeRepository must not assert negative release evidence for uncollected data", repo.Name)
		}
		if repo.ReleaseStatus != ReleaseStatusUnverified {
			t.Errorf("repo %q: expected ReleaseStatus=%q from raw API fixture, got %q", repo.Name, ReleaseStatusUnverified, repo.ReleaseStatus)
		}
		if repo.HasReleases {
			t.Errorf("repo %q: HasReleases must be false before bounded verification runs", repo.Name)
		}
	}
}

// TestNormalizeRepositoryFixturePreservesVerifiedDocumentationFlags (GI-TEST-001):
// Asserts that has_pages and has_wiki from the raw GitHub API JSON are correctly
// mapped to the normalized Repository struct and that HasDocs is derived from them,
// NOT from description/name substring heuristics.
func TestNormalizeRepositoryFixturePreservesVerifiedDocumentationFlags(t *testing.T) {
	rawRepos := loadRepositoryFixture(t, "real_octocat_repos.json")

	byName := make(map[string]Repository)
	for _, item := range rawRepos {
		if item.GetPrivate() {
			continue
		}
		repo, include := normalizePublicRepository(item)
		if include {
			byName[repo.Name] = repo
		}
	}

	// Spoon-Knife has has_pages: true, has_wiki: false, license: null
	spoon, ok := byName["Spoon-Knife"]
	if !ok {
		t.Fatal("fixture must contain Spoon-Knife")
	}
	if !spoon.HasPages {
		t.Errorf("Spoon-Knife: expected HasPages=true from raw API fixture has_pages:true")
	}
	if spoon.HasWiki {
		t.Errorf("Spoon-Knife: expected HasWiki=false from raw API fixture has_wiki:false")
	}
	if spoon.HasLicense {
		t.Errorf("Spoon-Knife: expected HasLicense=false from raw API fixture license:null")
	}
	if !spoon.HasDocs {
		t.Errorf("Spoon-Knife: expected HasDocs=true (derived from HasPages=true)")
	}

	// Hello-World has license: MIT, has_wiki: true
	hello, ok := byName["Hello-World"]
	if !ok {
		t.Fatal("fixture must contain Hello-World")
	}
	if !hello.HasLicense {
		t.Errorf("Hello-World: expected HasLicense=true from raw API fixture with MIT license")
	}
	if hello.License != "MIT License" {
		t.Errorf("Hello-World: expected License=%q, got %q", "MIT License", hello.License)
	}
	if !hello.HasWiki {
		t.Errorf("Hello-World: expected HasWiki=true from raw API fixture has_wiki:true")
	}
	if !hello.HasDocs {
		t.Errorf("Hello-World: expected HasDocs=true (derived from HasWiki=true and HasLicense=true)")
	}
}

// TestNormalizeRepositoryFixturePreservesForkAndArchivedFlags (GI-TEST-001):
// Asserts that Fork and Archived booleans are faithfully propagated from raw API JSON.
func TestNormalizeRepositoryFixturePreservesForkAndArchivedFlags(t *testing.T) {
	rawRepos := loadRepositoryFixture(t, "real_octocat_repos.json")

	byName := make(map[string]Repository)
	for _, item := range rawRepos {
		if item.GetPrivate() {
			continue
		}
		repo, include := normalizePublicRepository(item)
		if include {
			byName[repo.Name] = repo
		}
	}

	// boysenberry-repo-1 is fork:true in the fixture
	boysen, ok := byName["boysenberry-repo-1"]
	if !ok {
		t.Fatal("fixture must contain boysenberry-repo-1")
	}
	if !boysen.Fork {
		t.Errorf("boysenberry-repo-1: expected Fork=true from raw API fixture fork:true")
	}

	// git-consortium is archived:true in the fixture
	consortium, ok := byName["git-consortium"]
	if !ok {
		t.Fatal("fixture must contain git-consortium")
	}
	if !consortium.Archived {
		t.Errorf("git-consortium: expected Archived=true from raw API fixture archived:true")
	}
}

// TestNormalizeRepositoryFixtureHandlesNullLanguage (GI-TEST-001):
// GitHub API returns null for language when the repository has no detected language.
// normalizeRepository must treat null as an empty string, not panic or propagate "null".
func TestNormalizeRepositoryFixtureHandlesNullLanguage(t *testing.T) {
	rawRepos := loadRepositoryFixture(t, "real_octocat_repos.json")

	byName := make(map[string]Repository)
	for _, item := range rawRepos {
		if item.GetPrivate() {
			continue
		}
		repo, include := normalizePublicRepository(item)
		if include {
			byName[repo.Name] = repo
		}
	}

	// boysenberry-repo-1 has language: null in the fixture
	boysen, ok := byName["boysenberry-repo-1"]
	if !ok {
		t.Fatal("fixture must contain boysenberry-repo-1")
	}
	if boysen.Language != "" {
		t.Errorf("boysenberry-repo-1: expected Language='' for null API language, got %q", boysen.Language)
	}
}

// TestNormalizeProfileFixturePreservesPublicFields (GI-TEST-001):
// Feeds the real_octocat_profile.json fixture through the same normalization path
// as GetProfile, verifying that all public fields map correctly.
func TestNormalizeProfileFixturePreservesPublicFields(t *testing.T) {
	user := loadProfileFixture(t, "real_octocat_profile.json")

	profile := &UserProfile{
		Username:    user.GetLogin(),
		DisplayName: user.GetName(),
		AvatarURL:   user.GetAvatarURL(),
		Bio:         user.GetBio(),
		Location:    user.GetLocation(),
		Company:     user.GetCompany(),
		Blog:        user.GetBlog(),
		ProfileURL:  user.GetHTMLURL(),
		CreatedAt:   user.GetCreatedAt().Time,
		PublicRepos: user.GetPublicRepos(),
		Followers:   user.GetFollowers(),
		Following:   user.GetFollowing(),
	}

	if profile.Username != "octocat" {
		t.Errorf("expected Username=%q, got %q", "octocat", profile.Username)
	}
	if profile.DisplayName != "The Octocat" {
		t.Errorf("expected DisplayName=%q, got %q", "The Octocat", profile.DisplayName)
	}
	if profile.PublicRepos != 8 {
		t.Errorf("expected PublicRepos=8, got %d", profile.PublicRepos)
	}
	if profile.Followers != 16765 {
		t.Errorf("expected Followers=16765, got %d", profile.Followers)
	}
	if profile.Following != 9 {
		t.Errorf("expected Following=9, got %d", profile.Following)
	}
	if profile.CreatedAt.IsZero() {
		t.Errorf("expected non-zero CreatedAt from fixture created_at field")
	}
	if profile.CreatedAt.Year() != 2011 {
		t.Errorf("expected profile created in 2011, got %d", profile.CreatedAt.Year())
	}
	if profile.ProfileURL != "https://github.com/octocat" {
		t.Errorf("expected ProfileURL=%q, got %q", "https://github.com/octocat", profile.ProfileURL)
	}
}

// TestNormalizeRepositoryFixtureTimestampsAreParsed (GI-TEST-001):
// Verifies that ISO 8601 timestamps in the API fixture are parsed into non-zero
// time.Time values. A zero CreatedAt would indicate the timestamp field name
// is wrong or the JSON parser failed silently.
func TestNormalizeRepositoryFixtureTimestampsAreParsed(t *testing.T) {
	rawRepos := loadRepositoryFixture(t, "real_octocat_repos.json")

	for _, item := range rawRepos {
		if item.GetPrivate() {
			continue
		}
		repo, include := normalizePublicRepository(item)
		if !include {
			continue
		}
		if repo.CreatedAt.IsZero() {
			t.Errorf("repo %q: CreatedAt must be non-zero from raw API fixture created_at field", repo.Name)
		}
		if repo.CreatedAt.After(time.Now()) {
			t.Errorf("repo %q: CreatedAt %v is in the future, which is invalid", repo.Name, repo.CreatedAt)
		}
	}
}
