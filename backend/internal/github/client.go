package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	gh "github.com/google/go-github/v72/github"
	"golang.org/x/oauth2"
)

type ReadmeStatus string

const (
	ReadmeStatusVerifiedPresent ReadmeStatus = "verified_present"
	ReadmeStatusVerifiedAbsent  ReadmeStatus = "verified_absent"
	ReadmeStatusUnverified      ReadmeStatus = "unverified"
)

type ReleaseStatus string

const (
	ReleaseStatusVerifiedPresent ReleaseStatus = "verified_present"
	ReleaseStatusVerifiedAbsent  ReleaseStatus = "verified_absent"
	ReleaseStatusUnverified      ReleaseStatus = "unverified"
)

type TestingArtifactsStatus string

const (
	TestingArtifactsVerifiedPresent TestingArtifactsStatus = "verified_present"
	TestingArtifactsVerifiedAbsent  TestingArtifactsStatus = "verified_absent"
	TestingArtifactsUnverified      TestingArtifactsStatus = "unverified"
)

type ContributorsStatus string

const (
	ContributorsVerified   ContributorsStatus = "verified"
	ContributorsUnverified ContributorsStatus = "unverified"
)

type PullRequestCountStatus string

const (
	PullRequestCountsVerified   PullRequestCountStatus = "verified"
	PullRequestCountsUnverified PullRequestCountStatus = "unverified"
)

type RateLimitInfo struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

type UserProfile struct {
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Bio         string    `json:"bio"`
	Location    string    `json:"location"`
	Company     string    `json:"company"`
	Blog        string    `json:"blog"`
	ProfileURL  string    `json:"profile_url"`
	CreatedAt   time.Time `json:"created_at"`
	PublicRepos int       `json:"public_repos"`
	Followers   int       `json:"followers"`
	Following   int       `json:"following"`
}

type Repository struct {
	Name                           string                 `json:"name"`
	FullName                       string                 `json:"full_name"`
	Description                    string                 `json:"description"`
	URL                            string                 `json:"url"`
	Homepage                       string                 `json:"homepage"`
	Stars                          int                    `json:"stars"`
	Forks                          int                    `json:"forks"`
	Watchers                       int                    `json:"watchers"`
	OpenIssues                     int                    `json:"open_issues"`
	OpenIssuesCount                int                    `json:"open_issues_count"`
	OpenPullRequests               int                    `json:"open_pull_requests"`
	PullRequestCountStatus         PullRequestCountStatus `json:"pull_request_count_status"`
	PullRequestCountCheckAttempted bool                   `json:"pull_request_count_check_attempted"`
	Language                       string                 `json:"language"`
	LanguageBytes                  map[string]int         `json:"-"`
	TestingStatus                  TestingArtifactsStatus `json:"testing_artifacts_status"`
	TestingArtifacts               []string               `json:"testing_artifacts,omitempty"`
	TestingCheckAttempted          bool                   `json:"testing_artifacts_check_attempted"`
	ContributorsCount              int                    `json:"contributors_count"`
	ContributorsStatus             ContributorsStatus     `json:"contributors_status"`
	ContributorsCheckAttempted     bool                   `json:"contributors_check_attempted"`
	Topics                         []string               `json:"topics"`
	CreatedAt                      time.Time              `json:"created_at"`
	UpdatedAt                      time.Time              `json:"updated_at"`
	PushedAt                       time.Time              `json:"pushed_at"`
	DefaultBranch                  string                 `json:"default_branch"`
	License                        string                 `json:"license"`
	Archived                       bool                   `json:"archived"`
	Fork                           bool                   `json:"fork"`
	Visibility                     string                 `json:"visibility"`
	HasDocs                        bool                   `json:"has_docs"`
	HasPages                       bool                   `json:"has_pages"`
	HasWiki                        bool                   `json:"has_wiki"`
	HasReadme                      bool                   `json:"has_readme"`
	ReadmeStatus                   ReadmeStatus           `json:"readme_status"`
	ReadmeCheckAttempted           bool                   `json:"readme_check_attempted"`
	HasLicense                     bool                   `json:"has_license"`
	HasReleases                    bool                   `json:"has_releases"`
	ReleaseStatus                  ReleaseStatus          `json:"release_status"`
	ReleaseCount                   int                    `json:"release_count"`
	ReleaseCheckAttempted          bool                   `json:"release_check_attempted"`
}

type Client struct {
	client                *gh.Client
	token                 string
	repositoriesTruncated bool
}

func NewClientFromToken(token string) *Client {
	baseHTTPClient := &http.Client{Timeout: 20 * time.Second}
	if trimmed := strings.TrimSpace(token); trimmed != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: trimmed})
		baseHTTPClient = oauth2.NewClient(context.Background(), ts)
		baseHTTPClient.Timeout = 20 * time.Second
	}
	baseHTTPClient = withETagCache(baseHTTPClient, token)
	return &Client{client: gh.NewClient(baseHTTPClient), token: token}
}

func NewClientWithBaseURL(httpClient *http.Client, baseURL string, token string) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	if trimmed := strings.TrimSpace(token); trimmed != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: trimmed})
		httpClient = oauth2.NewClient(context.Background(), ts)
		httpClient.Timeout = 20 * time.Second
	}
	httpClient = withETagCache(httpClient, token)
	ghClient := gh.NewClient(httpClient)
	if baseURL != "" {
		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}
		u, err := url.Parse(baseURL)
		if err != nil {
			return nil, err
		}
		ghClient.BaseURL = u
	}
	return &Client{client: ghClient, token: token}, nil
}

func (c *Client) GetProfile(ctx context.Context, username string) (*UserProfile, error) {
	user, _, err := c.client.Users.Get(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return &UserProfile{
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
	}, nil
}

func (c *Client) GetRepositories(ctx context.Context, username string) ([]Repository, error) {
	opt := repositoryListOptions()
	var all []Repository
	c.repositoriesTruncated = false
	for {
		repos, resp, err := c.client.Repositories.ListByUser(ctx, username, opt)
		if err != nil {
			return nil, err
		}
		for _, item := range repos {
			repository, include := normalizePublicRepository(item)
			if !include {
				continue
			}
			all = append(all, repository)
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		if opt.Page >= maxRepositoryPages {
			c.repositoriesTruncated = true
			break
		}
		opt.Page = resp.NextPage
	}
	return all, nil
}

const maxRepositoryPages = 3

func (c *Client) RepositoriesTruncated() bool {
	return c.repositoriesTruncated
}

// PopulateLanguageBreakdown gathers language byte counts for the highest-signal repositories.
func (c *Client) PopulateLanguageBreakdown(ctx context.Context, repos []Repository, maxCandidates int) ([]Repository, error) {
	if len(repos) == 0 || maxCandidates <= 0 {
		return repos, nil
	}
	result := make([]Repository, len(repos))
	copy(result, repos)
	candidates := make([]int, len(result))
	for i := range candidates {
		candidates[i] = i
	}
	sort.Slice(candidates, func(i, j int) bool {
		left, right := result[candidates[i]], result[candidates[j]]
		leftScore, rightScore := candidateScore(left), candidateScore(right)
		if leftScore == rightScore {
			return left.FullName < right.FullName
		}
		return leftScore > rightScore
	})
	if len(candidates) > maxCandidates {
		candidates = candidates[:maxCandidates]
	}
	for _, index := range candidates {
		repo := result[index]
		parts := strings.SplitN(repo.FullName, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("repository %q has an invalid full name", repo.FullName)
		}
		langs, _, err := c.client.Repositories.ListLanguages(ctx, parts[0], parts[1])
		if err != nil {
			return nil, fmt.Errorf("fetch language breakdown for %s: %w", repo.FullName, err)
		}
		result[index].LanguageBytes = map[string]int{}
		for language, count := range langs {
			if strings.TrimSpace(language) == "" || count <= 0 {
				continue
			}
			result[index].LanguageBytes[language] = count
		}
	}
	return result, nil
}

func (c *Client) VerifyTestingArtifacts(ctx context.Context, repos []Repository, maxCandidates int) ([]Repository, error) {
	result := make([]Repository, len(repos))
	copy(result, repos)
	candidates := make([]int, 0, len(result))
	for i := range result {
		if result[i].TestingStatus == "" {
			result[i].TestingStatus = TestingArtifactsUnverified
		}
		if !result[i].Fork {
			candidates = append(candidates, i)
		}
	}
	if maxCandidates <= 0 {
		return result, nil
	}
	if maxCandidates > 5 {
		maxCandidates = 5
	}
	sort.Slice(candidates, func(i, j int) bool {
		left, right := result[candidates[i]], result[candidates[j]]
		leftScore, rightScore := candidateScore(left), candidateScore(right)
		if leftScore == rightScore {
			return left.FullName < right.FullName
		}
		return leftScore > rightScore
	})
	if maxCandidates < len(candidates) {
		candidates = candidates[:maxCandidates]
	}
	for _, index := range candidates {
		repo := result[index]
		parts := strings.SplitN(repo.FullName, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" || repo.DefaultBranch == "" {
			return nil, fmt.Errorf("repository %q is missing tree lookup metadata", repo.FullName)
		}
		result[index].TestingCheckAttempted = true
		tree, _, err := c.client.Git.GetTree(ctx, parts[0], parts[1], repo.DefaultBranch, true)
		if err != nil {
			return nil, fmt.Errorf("inspect test artifacts for %s: %w", repo.FullName, err)
		}
		if tree == nil {
			return nil, fmt.Errorf("inspect test artifacts for %s: GitHub returned an empty tree", repo.FullName)
		}
		artifacts := testingArtifactPaths(tree)
		result[index].TestingArtifacts = artifacts
		switch {
		case len(artifacts) > 0:
			result[index].TestingStatus = TestingArtifactsVerifiedPresent
		case tree.GetTruncated():
			result[index].TestingStatus = TestingArtifactsUnverified
		default:
			result[index].TestingStatus = TestingArtifactsVerifiedAbsent
		}
	}
	return result, nil
}

func (c *Client) VerifyContributorCounts(ctx context.Context, repos []Repository, maxCandidates int) ([]Repository, error) {
	result := make([]Repository, len(repos))
	copy(result, repos)
	candidates := make([]int, 0, len(result))
	for i := range result {
		if result[i].ContributorsStatus == "" {
			result[i].ContributorsStatus = ContributorsUnverified
		}
		if !result[i].Fork {
			candidates = append(candidates, i)
		}
	}
	if maxCandidates <= 0 {
		return result, nil
	}
	if maxCandidates > 5 {
		maxCandidates = 5
	}
	sort.Slice(candidates, func(i, j int) bool {
		left, right := result[candidates[i]], result[candidates[j]]
		leftScore, rightScore := candidateScore(left), candidateScore(right)
		if leftScore == rightScore {
			return left.FullName < right.FullName
		}
		return leftScore > rightScore
	})
	if len(candidates) > maxCandidates {
		candidates = candidates[:maxCandidates]
	}
	for _, index := range candidates {
		repo := result[index]
		parts := strings.SplitN(repo.FullName, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("repository %q has an invalid full name", repo.FullName)
		}
		result[index].ContributorsCheckAttempted = true
		contributors, response, err := c.client.Repositories.ListContributors(ctx, parts[0], parts[1], &gh.ListContributorsOptions{
			ListOptions: gh.ListOptions{PerPage: 1},
		})
		if err != nil {
			return nil, fmt.Errorf("fetch contributor count for %s: %w", repo.FullName, err)
		}
		count := len(contributors)
		if response != nil && response.LastPage > count {
			count = response.LastPage
		}
		result[index].ContributorsCount = count
		result[index].ContributorsStatus = ContributorsVerified
	}
	return result, nil
}

func (c *Client) VerifyOpenIssuePullRequestCounts(ctx context.Context, repos []Repository, maxCandidates int) ([]Repository, error) {
	result := make([]Repository, len(repos))
	copy(result, repos)
	candidates := make([]int, 0, len(result))
	for i := range result {
		if result[i].PullRequestCountStatus == "" {
			result[i].PullRequestCountStatus = PullRequestCountsUnverified
		}
		if !result[i].Fork {
			candidates = append(candidates, i)
		}
	}
	if maxCandidates <= 0 {
		return result, nil
	}
	if maxCandidates > 5 {
		maxCandidates = 5
	}
	sort.Slice(candidates, func(i, j int) bool {
		left, right := result[candidates[i]], result[candidates[j]]
		leftScore, rightScore := candidateScore(left), candidateScore(right)
		if leftScore == rightScore {
			return left.FullName < right.FullName
		}
		return leftScore > rightScore
	})
	if len(candidates) > maxCandidates {
		candidates = candidates[:maxCandidates]
	}
	for _, index := range candidates {
		repo := result[index]
		parts := strings.SplitN(repo.FullName, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("repository %q has an invalid full name", repo.FullName)
		}
		result[index].PullRequestCountCheckAttempted = true
		pullRequests, response, err := c.client.PullRequests.List(ctx, parts[0], parts[1], &gh.PullRequestListOptions{
			State:       "open",
			ListOptions: gh.ListOptions{PerPage: 1},
		})
		if err != nil {
			return nil, fmt.Errorf("fetch open pull request count for %s: %w", repo.FullName, err)
		}
		pullRequestCount := len(pullRequests)
		if response != nil && response.LastPage > pullRequestCount {
			pullRequestCount = response.LastPage
		}
		if pullRequestCount > repo.OpenIssues {
			return nil, fmt.Errorf("open pull request count for %s exceeds the combined open issue and pull request count", repo.FullName)
		}
		result[index].OpenPullRequests = pullRequestCount
		result[index].OpenIssuesCount = repo.OpenIssues - pullRequestCount
		result[index].PullRequestCountStatus = PullRequestCountsVerified
	}
	return result, nil
}

func testingArtifactPaths(tree *gh.Tree) []string {
	const maxArtifactsPerRepository = 8
	seen := make(map[string]struct{})
	artifacts := make([]string, 0, maxArtifactsPerRepository)
	add := func(path string) {
		if len(artifacts) >= maxArtifactsPerRepository {
			return
		}
		if _, exists := seen[path]; exists {
			return
		}
		seen[path] = struct{}{}
		artifacts = append(artifacts, path)
	}
	for _, entry := range tree.Entries {
		path := entry.GetPath()
		lowerPath := strings.ToLower(path)
		base := strings.ToLower(path[strings.LastIndex(path, "/")+1:])
		segments := strings.Split(lowerPath, "/")
		for _, segment := range segments {
			if entry.GetType() == "tree" && (segment == "test" || segment == "tests" || segment == "spec" || segment == "specs" || segment == "__tests__") {
				add(path)
				break
			}
		}
		if isTestArtifactFile(base) || isTestWorkflow(lowerPath) {
			add(path)
		}
	}
	return artifacts
}

func isTestArtifactFile(name string) bool {
	return strings.HasSuffix(name, "_test.go") ||
		strings.HasSuffix(name, "_test.py") ||
		strings.HasPrefix(name, "test_") && strings.HasSuffix(name, ".py") ||
		strings.Contains(name, ".test.") ||
		strings.Contains(name, ".spec.") ||
		strings.HasPrefix(name, "jest.config.") ||
		strings.HasPrefix(name, "vitest.config.") ||
		name == "pytest.ini" ||
		name == "tox.ini" ||
		name == "pom.xml" ||
		name == "go.mod" ||
		name == "phpunit.xml" ||
		name == ".rspec"
}

func isTestWorkflow(path string) bool {
	if !strings.HasPrefix(path, ".github/workflows/") {
		return false
	}
	name := strings.ToLower(path[strings.LastIndex(path, "/")+1:])
	if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
		return false
	}
	return strings.Contains(name, "test") || strings.Contains(name, "ci")
}

func repositoryListOptions() *gh.RepositoryListByUserOptions {
	return &gh.RepositoryListByUserOptions{Type: "owner", Sort: "updated", Direction: "desc", ListOptions: gh.ListOptions{PerPage: 100}}
}

func normalizePublicRepository(item *gh.Repository) (Repository, bool) {
	if item == nil || item.GetPrivate() {
		return Repository{}, false
	}
	return normalizeRepository(item), true
}

func (c *Client) RateLimit(ctx context.Context) (RateLimitInfo, error) {
	limits, _, err := c.client.RateLimits(ctx)
	if err != nil {
		return RateLimitInfo{}, err
	}
	if limits == nil || limits.Core == nil {
		return RateLimitInfo{}, nil
	}
	return RateLimitInfo{
		Limit:     limits.Core.Limit,
		Remaining: limits.Core.Remaining,
		Reset:     limits.Core.Reset.Time,
	}, nil
}

// CheckReadmePresence checks if a repository has a README using a lightweight HTTP HEAD request.
// Returns (true, nil) on HTTP 200, (false, nil) on HTTP 404, or (false, error) on other errors.
func (c *Client) CheckReadmePresence(ctx context.Context, owner, repo string) (bool, error) {
	req, err := c.client.NewRequest(http.MethodHead, fmt.Sprintf("repos/%s/%s/readme", owner, repo), nil)
	if err != nil {
		return false, err
	}
	resp, err := c.client.Do(ctx, req, nil)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		var errResp *gh.ErrorResponse
		if errors.As(err, &errResp) && errResp.Response != nil && errResp.Response.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	if resp != nil && resp.StatusCode == http.StatusOK {
		return true, nil
	}
	return false, nil
}

func candidateScore(repo Repository) float64 {
	score := float64(repo.Stars)*2 + float64(repo.Forks) + float64(repo.Watchers)/2
	if repo.HasLicense {
		score += 15
	}
	if !repo.Archived {
		score += 20
	}
	if !repo.Fork {
		score += 10
	}
	if repo.UpdatedAt.After(time.Now().AddDate(0, -6, 0)) {
		score += 25
	}
	return score
}

// VerifyCandidateReadmes performs bounded README verification for up to maxCandidates top repositories.
// Top candidates are determined by candidate score (stars, forks, recency, original status).
// Verified repositories have ReadmeStatus set to ReadmeStatusVerifiedPresent or ReadmeStatusVerifiedAbsent,
// and HasReadme set accordingly. Repositories not checked retain ReadmeStatusUnverified.
func (c *Client) VerifyCandidateReadmes(ctx context.Context, repos []Repository, maxCandidates int) []Repository {
	if len(repos) == 0 || maxCandidates <= 0 {
		return repos
	}

	type indexedCandidate struct {
		index int
		score float64
	}

	candidates := make([]indexedCandidate, len(repos))
	for i, r := range repos {
		candidates[i] = indexedCandidate{index: i, score: candidateScore(r)}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	limit := maxCandidates
	if len(candidates) < limit {
		limit = len(candidates)
	}

	result := make([]Repository, len(repos))
	copy(result, repos)
	for i := range result {
		if result[i].ReadmeStatus == "" {
			result[i].ReadmeStatus = ReadmeStatusUnverified
		}
	}

	for i := 0; i < limit; i++ {
		idx := candidates[i].index
		target := result[idx]

		owner := ""
		repoName := ""
		if strings.Contains(target.FullName, "/") {
			parts := strings.SplitN(target.FullName, "/", 2)
			owner = parts[0]
			repoName = parts[1]
		} else if target.Name != "" {
			repoName = target.Name
		}

		if owner == "" || repoName == "" {
			continue
		}

		result[idx].ReadmeCheckAttempted = true
		present, err := c.CheckReadmePresence(ctx, owner, repoName)
		if err != nil {
			result[idx].ReadmeStatus = ReadmeStatusUnverified
			result[idx].HasReadme = false
			continue
		}

		if present {
			result[idx].ReadmeStatus = ReadmeStatusVerifiedPresent
			result[idx].HasReadme = true
		} else {
			result[idx].ReadmeStatus = ReadmeStatusVerifiedAbsent
			result[idx].HasReadme = false
		}
	}

	return result
}

// CheckReleasePresence queries GitHub for repository releases.
// Returns (present bool, count int, err error).
// Returns (true, count, nil) if releases exist, (false, 0, nil) on HTTP 404 or empty releases,
// or (false, 0, err) on network/rate-limit/API errors.
func (c *Client) CheckReleasePresence(ctx context.Context, owner, repo string) (bool, int, error) {
	releases, resp, err := c.client.Repositories.ListReleases(ctx, owner, repo, &gh.ListOptions{PerPage: 10})
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, 0, nil
		}
		var errResp *gh.ErrorResponse
		if errors.As(err, &errResp) && errResp.Response != nil && errResp.Response.StatusCode == http.StatusNotFound {
			return false, 0, nil
		}
		return false, 0, err
	}
	count := len(releases)
	if count > 0 {
		return true, count, nil
	}
	return false, 0, nil
}

// VerifyCandidateReleases performs bounded release verification for up to maxCandidates top active original repositories.
// Candidates must be original (!repo.Fork), non-archived (!repo.Archived), and updated/pushed within the last 180 days.
// Candidates are ranked by candidate score. Verified repositories have ReleaseStatus set to ReleaseStatusVerifiedPresent
// or ReleaseStatusVerifiedAbsent, HasReleases set accordingly, and ReleaseCount populated.
// Repositories not checked retain ReleaseStatusUnverified.
func (c *Client) VerifyCandidateReleases(ctx context.Context, repos []Repository, maxCandidates int) []Repository {
	if len(repos) == 0 || maxCandidates <= 0 {
		return repos
	}

	result := make([]Repository, len(repos))
	copy(result, repos)
	for i := range result {
		if result[i].ReleaseStatus == "" {
			result[i].ReleaseStatus = ReleaseStatusUnverified
		}
	}

	cutoff := time.Now().AddDate(0, 0, -180)
	type indexedCandidate struct {
		index int
		score float64
	}

	var candidates []indexedCandidate
	for i, r := range result {
		if r.Fork || r.Archived {
			continue
		}
		if r.PushedAt.Before(cutoff) && r.UpdatedAt.Before(cutoff) {
			continue
		}
		candidates = append(candidates, indexedCandidate{index: i, score: candidateScore(r)})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	limit := maxCandidates
	if len(candidates) < limit {
		limit = len(candidates)
	}

	for i := 0; i < limit; i++ {
		idx := candidates[i].index
		target := result[idx]

		owner := ""
		repoName := ""
		if strings.Contains(target.FullName, "/") {
			parts := strings.SplitN(target.FullName, "/", 2)
			owner = parts[0]
			repoName = parts[1]
		} else if target.Name != "" {
			repoName = target.Name
		}

		if owner == "" || repoName == "" {
			continue
		}

		result[idx].ReleaseCheckAttempted = true
		present, count, err := c.CheckReleasePresence(ctx, owner, repoName)
		if err != nil {
			result[idx].ReleaseStatus = ReleaseStatusUnverified
			result[idx].HasReleases = false
			result[idx].ReleaseCount = 0
			continue
		}

		if present {
			result[idx].ReleaseStatus = ReleaseStatusVerifiedPresent
			result[idx].HasReleases = true
			result[idx].ReleaseCount = count
		} else {
			result[idx].ReleaseStatus = ReleaseStatusVerifiedAbsent
			result[idx].HasReleases = false
			result[idx].ReleaseCount = 0
		}
	}

	return result
}

func normalizeRepository(item *gh.Repository) Repository {
	repo := Repository{
		Name:                   item.GetName(),
		FullName:               item.GetFullName(),
		Description:            item.GetDescription(),
		URL:                    item.GetHTMLURL(),
		Homepage:               item.GetHomepage(),
		Stars:                  item.GetStargazersCount(),
		Forks:                  item.GetForksCount(),
		Watchers:               item.GetWatchersCount(),
		OpenIssues:             item.GetOpenIssuesCount(),
		Language:               item.GetLanguage(),
		TestingStatus:          TestingArtifactsUnverified,
		ContributorsStatus:     ContributorsUnverified,
		PullRequestCountStatus: PullRequestCountsUnverified,
		Topics:                 item.Topics,
		CreatedAt:              item.GetCreatedAt().Time,
		UpdatedAt:              item.GetUpdatedAt().Time,
		PushedAt:               item.GetPushedAt().Time,
		DefaultBranch:          item.GetDefaultBranch(),
		Archived:               item.GetArchived(),
		Fork:                   item.GetFork(),
		Visibility:             item.GetVisibility(),
		HasPages:               item.GetHasPages(),
		HasWiki:                item.GetHasWiki(),
	}
	if item.GetLicense() != nil {
		repo.License = item.GetLicense().GetName()
		repo.HasLicense = true
	}
	repo.ReadmeStatus = ReadmeStatusUnverified
	repo.HasReadme = false
	repo.ReleaseStatus = ReleaseStatusUnverified
	repo.ReleaseCount = 0
	repo.HasReleases = false
	repo.HasDocs = repo.HasPages || repo.HasWiki || repo.HasLicense
	return repo
}
