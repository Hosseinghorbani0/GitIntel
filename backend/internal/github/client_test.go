package github

import (
	"testing"

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

func stringPointer(value string) *string { return &value }
