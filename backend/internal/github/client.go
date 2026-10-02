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
	Name          string       `json:"name"`
	FullName      string       `json:"full_name"`
	Description   string       `json:"description"`
	URL           string       `json:"url"`
	Homepage      string       `json:"homepage"`
	Stars         int          `json:"stars"`
	Forks         int          `json:"forks"`
	Watchers      int          `json:"watchers"`
	OpenIssues    int          `json:"open_issues"`
	Language      string       `json:"language"`
	Topics        []string     `json:"topics"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	PushedAt      time.Time    `json:"pushed_at"`
	DefaultBranch string       `json:"default_branch"`
	License       string       `json:"license"`
	Archived      bool         `json:"archived"`
	Fork          bool         `json:"fork"`
	Visibility    string       `json:"visibility"`
	HasDocs       bool         `json:"has_docs"`
	HasReadme     bool         `json:"has_readme"`
	ReadmeStatus  ReadmeStatus `json:"readme_status"`
	HasLicense    bool         `json:"has_license"`
	HasReleases   bool         `json:"has_releases"`
}

type Client struct {
	client *gh.Client
	token  string
}

func NewClientFromToken(token string) *Client {
	baseHTTPClient := &http.Client{Timeout: 20 * time.Second}
	if trimmed := strings.TrimSpace(token); trimmed != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: trimmed})
		baseHTTPClient = oauth2.NewClient(context.Background(), ts)
		baseHTTPClient.Timeout = 20 * time.Second
	}
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
		opt.Page = resp.NextPage
	}
	return all, nil
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

func normalizeRepository(item *gh.Repository) Repository {
	repo := Repository{
		Name:          item.GetName(),
		FullName:      item.GetFullName(),
		Description:   item.GetDescription(),
		URL:           item.GetHTMLURL(),
		Homepage:      item.GetHomepage(),
		Stars:         item.GetStargazersCount(),
		Forks:         item.GetForksCount(),
		Watchers:      item.GetWatchersCount(),
		OpenIssues:    item.GetOpenIssuesCount(),
		Language:      item.GetLanguage(),
		Topics:        item.Topics,
		CreatedAt:     item.GetCreatedAt().Time,
		UpdatedAt:     item.GetUpdatedAt().Time,
		PushedAt:      item.GetPushedAt().Time,
		DefaultBranch: item.GetDefaultBranch(),
		Archived:      item.GetArchived(),
		Fork:          item.GetFork(),
		Visibility:    item.GetVisibility(),
	}
	if item.GetLicense() != nil {
		repo.License = item.GetLicense().GetName()
		repo.HasLicense = true
	}
	description := strings.ToLower(repo.Description)
	homepage := strings.ToLower(repo.Homepage)
	repo.ReadmeStatus = ReadmeStatusUnverified
	repo.HasReadme = false
	repo.HasDocs = strings.Contains(description, "docs") || strings.Contains(homepage, "docs") || strings.Contains(strings.ToLower(repo.Name), "docs")
	return repo
}

