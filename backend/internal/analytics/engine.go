package analytics

import (
	"fmt"
	"sort"
	"strings"
	"time"

	gh "gitintel/backend/internal/github"
)

type LanguageBucket struct {
	Name       string  `json:"name"`
	Percentage float64 `json:"percentage"`
	Count      int     `json:"count"`
}

type ActivityMetrics struct {
	Commits           string `json:"commits"`
	PullRequests      string `json:"pull_requests"`
	Issues            string `json:"issues"`
	Releases          string `json:"releases"`
	RepositoryUpdates string `json:"repository_updates"`
}

type Signal struct {
	Name           string             `json:"name"`
	Level          string             `json:"level"`
	Classification string             `json:"classification"`
	Confidence     string             `json:"confidence"`
	Weight         float64            `json:"weight"`
	Value          float64            `json:"value"`
	Metric         map[string]float64 `json:"metric"`
	Evidence       []string           `json:"evidence"`
	Repositories   []string           `json:"repositories,omitempty"`
	Limitations    []string           `json:"limitations"`
	Explanation    string             `json:"explanation"`
}

type FeaturedProject struct {
	Repository gh.Repository `json:"repository"`
	Why        []string      `json:"why"`
}

type Analysis struct {
	Profile              gh.UserProfile         `json:"profile"`
	Metrics              map[string]interface{} `json:"metrics"`
	Summary              string                 `json:"summary"`
	LanguageDistribution []LanguageBucket       `json:"language_distribution"`
	LanguageBasis        string                 `json:"language_distribution_basis"`
	LanguageSampleSize   int                    `json:"language_repositories_sampled"`
	Activity             ActivityMetrics        `json:"activity"`
	Signals              []Signal               `json:"signals"`
	FeaturedProjects     []FeaturedProject      `json:"featured_projects"`
	RepositoryLinks      []string               `json:"repository_links"`
	ResumeMarkdown        string                 `json:"resume_markdown"`
	PortfolioMarkdown     string                 `json:"portfolio_markdown"`
	ReportMarkdown        string                 `json:"report_markdown"`
	ProfileReadmeMarkdown string                 `json:"profile_readme_markdown"`
	AIPromptMarkdown      string                 `json:"ai_prompt_markdown"`
}

type Engine struct{}

func NewEngine() *Engine { return &Engine{} }

func (e *Engine) Analyze(profile gh.UserProfile, repos []gh.Repository) Analysis {
	metrics := map[string]interface{}{
		"total_repositories":        len(repos),
		"original_repositories":     countOriginal(repos),
		"forked_repositories":       countForked(repos),
		"archived_repositories":     countArchived(repos),
		"total_stars":               sum(repos, func(r gh.Repository) int { return r.Stars }),
		"total_forks":               sum(repos, func(r gh.Repository) int { return r.Forks }),
		"total_open_issues_and_prs": sum(repos, func(r gh.Repository) int { return r.OpenIssues }),
	}
	langs := buildLanguageDistribution(repos)
	languageBasis, languageSampleSize := languageDistributionSource(repos)
	signals := buildSignals(repos)
	featured := selectFeatured(repos)
	report := buildReportMarkdown(profile, repos, langs, languageBasis, languageSampleSize, signals, featured)
	analysis := Analysis{
		Profile:               profile,
		Metrics:               metrics,
		Summary:               buildSummary(profile, repos, langs, languageBasis, languageSampleSize),
		LanguageDistribution:  langs,
		LanguageBasis:         languageBasis,
		LanguageSampleSize:    languageSampleSize,
		Activity:              buildActivityMetrics(repos),
		Signals:               signals,
		FeaturedProjects:      featured,
		RepositoryLinks:       buildRepositoryLinks(featured),
		ResumeMarkdown:        buildResumeMarkdown(profile, featured),
		PortfolioMarkdown:     buildPortfolioMarkdown(profile, featured, signals),
		ReportMarkdown:        report,
		ProfileReadmeMarkdown: buildProfileReadmeMarkdown(profile, featured, signals, langs, languageBasis),
		AIPromptMarkdown:      buildAIPromptMarkdown(profile, featured, signals, langs),
	}
	return analysis
}

func buildSummary(profile gh.UserProfile, repos []gh.Repository, langs []LanguageBucket, languageBasis string, languageSampleSize int) string {
	primary := "General-purpose full-stack activity"
	if len(langs) > 0 {
		primary = langs[0].Name
	}
	if len(repos) == 0 {
		return "No public repositories were returned for this profile, so repository-based engineering signals cannot be assessed."
	}
	maintained := countRecentlyUpdated(repos, 180)
	basisDescription := "repository frequency (a repository-count estimate, not code volume)"
	if languageBasis == "code_bytes_sampled" {
		basisDescription = fmt.Sprintf("code bytes sampled from %d repositories", languageSampleSize)
	}
	return fmt.Sprintf("This GitHub profile contains %d public repositories. The largest language share in the available distribution is %s, based on %s. %d repositories have update timestamps within the last 180 days. These metadata signals do not establish code quality, contribution content, or engineering competence.", len(repos), primary, basisDescription, maintained)
}

func buildLanguageDistribution(repos []gh.Repository) []LanguageBucket {
	bytesByLanguage := map[string]int{}
	fallbackCounts := map[string]int{}
	hasByteData := false
	for _, repo := range repos {
		if repo.LanguageBytes != nil {
			hasByteData = true
			for language, count := range repo.LanguageBytes {
				if strings.TrimSpace(language) == "" || count <= 0 {
					continue
				}
				bytesByLanguage[language] += count
			}
			continue
		}
		if strings.TrimSpace(repo.Language) == "" {
			continue
		}
		fallbackCounts[repo.Language]++
	}
	weights := bytesByLanguage
	if !hasByteData {
		weights = fallbackCounts
	}
	if len(weights) == 0 {
		return nil
	}
	var buckets []LanguageBucket
	for name, count := range weights {
		buckets = append(buckets, LanguageBucket{Name: name, Count: count})
	}
	sort.Slice(buckets, func(i, j int) bool { return buckets[i].Count > buckets[j].Count })
	total := 0
	for _, bucket := range buckets {
		total += bucket.Count
	}
	for i := range buckets {
		if total == 0 {
			buckets[i].Percentage = 0
			continue
		}
		buckets[i].Percentage = float64(buckets[i].Count) / float64(total) * 100
	}
	return buckets
}

func languageDistributionSource(repos []gh.Repository) (string, int) {
	sampled := 0
	for _, repo := range repos {
		if repo.LanguageBytes != nil {
			sampled++
		}
	}
	if sampled > 0 {
		return "code_bytes_sampled", sampled
	}
	return "repository_frequency_fallback", 0
}

func buildActivityMetrics(repos []gh.Repository) ActivityMetrics {
	return ActivityMetrics{
		Commits:           "Not available",
		PullRequests:      "Not available",
		Issues:            "Not available",
		Releases:          fmt.Sprintf("%d repositories with release metadata", countWithReleases(repos)),
		RepositoryUpdates: fmt.Sprintf("%d repositories updated in the last 180 days", countRecentlyUpdated(repos, 180)),
	}
}

func buildSignals(repos []gh.Repository) []Signal {
	var signals []Signal
	signals = append(signals, computeCodeEvolution(repos))
	signals = append(signals, computeMaintenance(repos))
	signals = append(signals, computeTestingEvidence(repos))
	signals = append(signals, computeCollaboration(repos))
	signals = append(signals, computeDocumentation(repos))
	signals = append(signals, computeReleaseActivity(repos))
	return signals
}

func computeCodeEvolution(repos []gh.Repository) Signal {
	updated90 := countRecentlyUpdated(repos, 90)
	updated180 := countRecentlyUpdated(repos, 180)
	var evidenceRepositories []string
	for _, repo := range repos {
		if repo.UpdatedAt.After(time.Now().AddDate(0, 0, -180)) || repo.PushedAt.After(time.Now().AddDate(0, 0, -180)) {
			if identifier := repositoryIdentifier(repo); identifier != "" {
				evidenceRepositories = append(evidenceRepositories, identifier)
			}
		}
	}
	level := "Insufficient data"
	confidence := "Low"
	evidence := []string{"No recent repository update signals detected."}
	if updated180 > 0 {
		level = "Moderate"
		confidence = "Medium"
		evidence = []string{fmt.Sprintf("%d repositories have been updated in the last 180 days.", updated180)}
		if updated90 > 0 {
			level = "Strong"
			confidence = "Medium"
			evidence = []string{fmt.Sprintf("%d repositories show recent activity in the last 90 days.", updated90)}
		}
	}
	return Signal{Name: "Code Evolution", Level: level, Classification: level, Confidence: confidence, Weight: 25, Value: float64(updated90), Metric: map[string]float64{"repositories_updated_90d": float64(updated90), "repositories_updated_180d": float64(updated180), "repositories_total": float64(len(repos))}, Evidence: evidence, Repositories: evidenceRepositories, Limitations: []string{"Tracks recent push and repository update timestamps. Commit volume, diff size, and branch activity are not collected in this REST collection.", "A repository metadata edit can update timestamps without a corresponding code change; recency does not establish contribution volume or code quality."}, Explanation: "This signal reports repository timestamp recency, not a direct measure of code evolution or technical skill."}
}

func computeMaintenance(repos []gh.Repository) Signal {
	active := countRecentlyUpdated(repos, 180)
	archived := countArchived(repos)
	level := "Limited"
	classification := ""
	confidence := "Medium"
	evidence := []string{"Repository maintenance indicators are limited."}
	var evidenceRepositories []string
	for _, repo := range repos {
		if repo.Archived || repo.UpdatedAt.After(time.Now().AddDate(0, 0, -180)) || repo.PushedAt.After(time.Now().AddDate(0, 0, -180)) {
			if identifier := repositoryIdentifier(repo); identifier != "" {
				evidenceRepositories = append(evidenceRepositories, identifier)
			}
		}
	}
	if active > 0 {
		level = "Strong"
		confidence = "Medium"
		evidence = []string{fmt.Sprintf("%d repositories have recent updates and visible maintenance activity.", active)}
		if archived > 0 {
			level = "Moderate"
			evidence = []string{fmt.Sprintf("%d repositories are archived, while %d repositories remain actively maintained.", archived, active)}
		}
	} else if archived > 0 {
		classification = "Archived / Completed"
		evidence = []string{fmt.Sprintf("%d repositories are archived; inactivity alone is not treated as poor maintenance.", archived)}
	} else if len(repos) == 0 {
		level = "Insufficient data"
		classification = level
		confidence = "Low"
		evidence = []string{"No repositories were available to assess maintenance timestamps."}
	}
	if classification == "" {
		classification = level
	}
	return Signal{Name: "Maintenance", Level: level, Classification: classification, Confidence: confidence, Weight: 20, Value: float64(active), Metric: map[string]float64{"repositories_updated_180d": float64(active), "repositories_archived": float64(archived), "repositories_total": float64(len(repos))}, Evidence: evidence, Repositories: evidenceRepositories, Limitations: []string{"Uses update timestamps and archived flags only; inactivity is not evidence of poor engineering or project quality.", "Issue handling, commit contents, and maintenance quality were not inspected."}, Explanation: "Maintenance reports repository activity and archived lifecycle metadata; archived repositories are treated as potentially completed projects."}
}

func computeTestingEvidence(repos []gh.Repository) Signal {
	artifactRepositories := 0
	inspectedRepositories := 0
	var evidenceRepositories []string
	var observedArtifacts []string
	for _, repo := range repos {
		if repo.TestingStatus == gh.TestingArtifactsUnverified || repo.TestingStatus == "" {
			continue
		}
		inspectedRepositories++
		if repo.TestingStatus == gh.TestingArtifactsVerifiedPresent && len(repo.TestingArtifacts) > 0 {
			artifactRepositories++
			if identifier := repositoryIdentifier(repo); identifier != "" {
				evidenceRepositories = append(evidenceRepositories, identifier)
				for _, artifact := range repo.TestingArtifacts {
					if len(observedArtifacts) == 3 {
						break
					}
					observedArtifacts = append(observedArtifacts, fmt.Sprintf("%s: %s", identifier, artifact))
				}
			}
		}
	}
	level := "Insufficient data"
	confidence := "Low"
	evidence := []string{"Test artifact metadata was not collected for the analyzed repositories."}
	if artifactRepositories > 0 {
		level = "Moderate"
		confidence = "Medium"
		evidence = []string{fmt.Sprintf("Observable test files, test directories, or CI configuration were found in %d inspected repositories.", artifactRepositories)}
		evidence = append(evidence, observedArtifacts...)
		if artifactRepositories >= 3 {
			level = "Strong"
		}
	} else if inspectedRepositories > 0 {
		level = "Limited"
		confidence = "Medium"
		evidence = []string{fmt.Sprintf("No supported test artifacts were found in %d inspected repositories.", inspectedRepositories)}
	}
	return Signal{Name: "Testing Evidence", Level: level, Classification: level, Confidence: confidence, Weight: 20, Value: float64(artifactRepositories), Metric: map[string]float64{"repositories_with_test_artifacts": float64(artifactRepositories), "repositories_test_artifacts_inspected": float64(inspectedRepositories), "repositories_total": float64(len(repos))}, Evidence: evidence, Repositories: evidenceRepositories, Limitations: []string{"Observed test directory structures and CI configuration; test coverage percentages and assertion rigor were not evaluated.", fmt.Sprintf("Test artifact trees were inspected for up to 5 original repositories; %d repositories remain unverified.", len(repos)-inspectedRepositories)}, Explanation: "This signal records visible test-related files, directories, and CI workflow names; it does not evaluate test quality or execution."}
}

func computeCollaboration(repos []gh.Repository) Signal {
	personalForks := countForked(repos)
	incomingForks := 0
	originalReposWithIncomingForks := 0
	issueAndPRCount := 0
	verifiedIssueCount := 0
	verifiedPullRequestCount := 0
	repositoriesWithVerifiedIssuePRCounts := 0
	contributorsObserved := 0
	contributorCountsVerified := 0
	repositoriesWithMultipleContributors := 0
	var evidenceRepositories []string
	seenEvidenceRepositories := make(map[string]struct{})
	addEvidenceRepository := func(repo gh.Repository) {
		if identifier := repositoryIdentifier(repo); identifier != "" {
			if _, exists := seenEvidenceRepositories[identifier]; !exists {
				seenEvidenceRepositories[identifier] = struct{}{}
				evidenceRepositories = append(evidenceRepositories, identifier)
			}
		}
	}
	for _, repo := range repos {
		if repo.Fork {
			continue
		}
		if repo.ContributorsStatus == gh.ContributorsVerified {
			contributorCountsVerified++
			contributorsObserved += repo.ContributorsCount
			if repo.ContributorsCount > 1 {
				repositoriesWithMultipleContributors++
				addEvidenceRepository(repo)
			}
		}
		incomingForks += repo.Forks
		issueAndPRCount += repo.OpenIssues
		if repo.PullRequestCountStatus == gh.PullRequestCountsVerified {
			repositoriesWithVerifiedIssuePRCounts++
			verifiedIssueCount += repo.OpenIssuesCount
			verifiedPullRequestCount += repo.OpenPullRequests
			if repo.OpenIssuesCount > 0 || repo.OpenPullRequests > 0 {
				addEvidenceRepository(repo)
			}
		}
		if repo.Forks > 0 {
			originalReposWithIncomingForks++
			addEvidenceRepository(repo)
		}
		if repo.OpenIssues > 0 && repo.Forks == 0 {
			addEvidenceRepository(repo)
		}
	}
	level := "Limited"
	confidence := "Low"
	evidence := []string{fmt.Sprintf("No repositories with multiple contributors were observed among %d original repositories with verified contributor counts (%d contributors observed).", contributorCountsVerified, contributorsObserved)}
	if incomingForks >= 2 || issueAndPRCount > 0 || repositoriesWithMultipleContributors > 0 {
		level = "Moderate"
		confidence = "Medium"
		evidence = []string{fmt.Sprintf("Original repositories show %d incoming forks and %d combined open issues and pull requests; bounded checks separated %d open issues from %d open pull requests across %d repositories, with %d repositories showing multiple observed contributors.", incomingForks, issueAndPRCount, verifiedIssueCount, verifiedPullRequestCount, repositoriesWithVerifiedIssuePRCounts, repositoriesWithMultipleContributors)}
	} else if personalForks > 0 || incomingForks > 0 {
		evidence = []string{fmt.Sprintf("%d personal forks and %d incoming forks were observed; a single fork alone is not treated as collaboration evidence.", personalForks, incomingForks)}
	}
	return Signal{Name: "Collaboration", Level: level, Classification: level, Confidence: confidence, Weight: 15, Value: float64(incomingForks), Metric: map[string]float64{"personal_forked_repositories": float64(personalForks), "incoming_forks_on_original_repositories": float64(incomingForks), "original_repositories_with_incoming_forks": float64(originalReposWithIncomingForks), "open_issues_and_prs": float64(issueAndPRCount), "verified_open_issues_on_bounded_sample": float64(verifiedIssueCount), "verified_open_pull_requests_on_bounded_sample": float64(verifiedPullRequestCount), "repositories_with_verified_issue_pr_counts": float64(repositoriesWithVerifiedIssuePRCounts), "repositories_with_verified_contributor_counts": float64(contributorCountsVerified), "repositories_with_multiple_contributors": float64(repositoriesWithMultipleContributors), "contributors_observed_on_bounded_sample": float64(contributorsObserved), "repositories_total": float64(len(repos))}, Evidence: evidence, Repositories: evidenceRepositories, Limitations: []string{"The repository-list open count combines issues and pull requests; separate issue and pull-request counts are available only for up to 5 high-signal original repositories.", "Contributor counts are collected only for up to 5 high-signal original repositories; external pull requests to non-owned repositories require GraphQL collection.", "Incoming forks indicate external copying, not necessarily direct collaboration; contributor identities and discussion quality are not evaluated."}, Explanation: "Collaboration is inferred conservatively from incoming forks, bounded contributor and issue/pull-request counts, and combined repository-list counts; personal forks are not counted as collaboration."}
}

func computeDocumentation(repos []gh.Repository) Signal {
	count := 0
	verifiedReadmeCount := 0
	unverifiedReadmeCount := 0
	pagesCount, wikiCount, licenseCount := 0, 0, 0
	var evidenceRepositories []string
	for _, repo := range repos {
		hasVerifiedReadme := repo.ReadmeStatus == gh.ReadmeStatusVerifiedPresent || repo.HasReadme
		if hasVerifiedReadme {
			verifiedReadmeCount++
		} else if repo.ReadmeStatus == gh.ReadmeStatusUnverified || repo.ReadmeStatus == "" {
			unverifiedReadmeCount++
		}
		if repo.HasPages {
			pagesCount++
		}
		if repo.HasWiki {
			wikiCount++
		}
		if repo.HasLicense {
			licenseCount++
		}
		if hasVerifiedReadme || repo.HasPages || repo.HasWiki || repo.HasLicense {
			count++
			if identifier := repositoryIdentifier(repo); identifier != "" {
				evidenceRepositories = append(evidenceRepositories, identifier)
			}
		}
	}
	level := "Limited"
	confidence := "Medium"
	var evidence []string
	if count > 0 {
		level = "Moderate"
		if verifiedReadmeCount > 0 {
			evidence = append(evidence, fmt.Sprintf("%d repositories have verified README documentation.", verifiedReadmeCount))
		}
		evidence = append(evidence, fmt.Sprintf("%d repositories have verified documentation artifacts (Pages: %d, Wiki: %d, LICENSE: %d).", count, pagesCount, wikiCount, licenseCount))
		if count*2 >= len(repos) {
			level = "Strong"
		}
	} else {
		evidence = []string{"No verified README, GitHub Pages, Wiki, or LICENSE artifacts were observed."}
	}

	limitations := []string{
		"Documentation is based only on verified README presence, GitHub Pages/Wiki metadata, and LICENSE metadata.",
	}
	if unverifiedReadmeCount > 0 {
		limitations = append(limitations, fmt.Sprintf("README presence was unverified for %d repositories due to rate-limit bounding.", unverifiedReadmeCount))
	}
	limitations = append(limitations, "README content, structure, and documentation quality were not fetched or inspected.")

	metrics := map[string]float64{
		"repositories_with_documentation_artifacts": float64(count),
		"repositories_with_verified_readme":         float64(verifiedReadmeCount),
		"repositories_with_unverified_readme":       float64(unverifiedReadmeCount),
		"repositories_with_github_pages":            float64(pagesCount),
		"repositories_with_github_wiki":             float64(wikiCount),
		"repositories_with_license":                 float64(licenseCount),
		"repositories_total":                        float64(len(repos)),
	}

	return Signal{
		Name:           "Documentation",
		Level:          level,
		Classification: level,
		Confidence:     confidence,
		Weight:         10,
		Value:          float64(count),
		Metric:         metrics,
		Evidence:       evidence,
		Repositories:   evidenceRepositories,
		Limitations:    limitations,
		Explanation:    "Documentation classification uses only verified README presence and GitHub Pages, Wiki, and LICENSE artifacts.",
	}
}

func computeReleaseActivity(repos []gh.Repository) Signal {
	verifiedPresentCount := 0
	verifiedAbsentCount := 0
	unverifiedCount := 0
	var releaseRepos []string
	var releaseReferences []string

	for _, repo := range repos {
		if repo.ReleaseStatus == gh.ReleaseStatusVerifiedPresent || repo.HasReleases {
			verifiedPresentCount++
			name := repo.Name
			if name == "" {
				name = repo.FullName
			}
			if name != "" {
				releaseRepos = append(releaseRepos, name)
			}
			if identifier := repositoryIdentifier(repo); identifier != "" {
				releaseReferences = append(releaseReferences, identifier)
			}
		} else if repo.ReleaseStatus == gh.ReleaseStatusVerifiedAbsent {
			verifiedAbsentCount++
		} else {
			unverifiedCount++
		}
	}

	level := "Insufficient data"
	confidence := "Low"
	var evidence []string

	if verifiedPresentCount >= 2 {
		level = "Strong"
		confidence = "Medium"
		evidence = []string{fmt.Sprintf("%d repositories have verified releases or deployment tags (%s).", verifiedPresentCount, strings.Join(releaseRepos, ", "))}
	} else if verifiedPresentCount == 1 {
		level = "Moderate"
		confidence = "Medium"
		evidence = []string{fmt.Sprintf("1 repository has verified releases or deployment tags (%s).", releaseRepos[0])}
	} else if verifiedAbsentCount > 0 && unverifiedCount == 0 {
		level = "Limited"
		confidence = "Medium"
		evidence = []string{"Verified repositories showed no published releases or deployment tags."}
	} else {
		evidence = []string{"No verified releases were observed across candidate repositories."}
	}

	limitations := []string{
		"Release verification is bounded to top active original repositories to preserve API rate limits.",
	}
	if unverifiedCount > 0 {
		limitations = append(limitations, fmt.Sprintf("Release status was unverified for %d repositories due to rate-limit bounding or age thresholds.", unverifiedCount))
	}
	limitations = append(limitations, "Assets, changelogs, and deployment verification pipelines were not evaluated.")

	metrics := map[string]float64{
		"repositories_with_verified_releases":  float64(verifiedPresentCount),
		"repositories_with_verified_absent":    float64(verifiedAbsentCount),
		"repositories_with_unverified_release": float64(unverifiedCount),
		"repositories_total":                   float64(len(repos)),
	}

	return Signal{
		Name:           "Release / Delivery Signals",
		Level:          level,
		Classification: level,
		Confidence:     confidence,
		Weight:         10,
		Value:          float64(verifiedPresentCount),
		Metric:         metrics,
		Evidence:       evidence,
		Repositories:   releaseReferences,
		Limitations:    limitations,
		Explanation:    "Release activity evaluates published GitHub releases on active original repositories.",
	}
}

func repositoryIdentifier(repo gh.Repository) string {
	if repo.FullName != "" {
		return repo.FullName
	}
	return repo.Name
}

func countOriginal(repos []gh.Repository) int {
	count := 0
	for _, repo := range repos {
		if !repo.Fork {
			count++
		}
	}
	return count
}

func countForked(repos []gh.Repository) int {
	count := 0
	for _, repo := range repos {
		if repo.Fork {
			count++
		}
	}
	return count
}
func countArchived(repos []gh.Repository) int {
	count := 0
	for _, repo := range repos {
		if repo.Archived {
			count++
		}
	}
	return count
}
func countWithReleases(repos []gh.Repository) int {
	count := 0
	for _, repo := range repos {
		if repo.ReleaseStatus == gh.ReleaseStatusVerifiedPresent || repo.HasReleases {
			count++
		}
	}
	return count
}
func countRecentlyUpdated(repos []gh.Repository, days int) int {
	cutoff := time.Now().AddDate(0, 0, -days)
	count := 0
	for _, repo := range repos {
		if repo.UpdatedAt.After(cutoff) || repo.PushedAt.After(cutoff) {
			count++
		}
	}
	return count
}

func sum(repos []gh.Repository, fn func(gh.Repository) int) int {
	total := 0
	for _, repo := range repos {
		total += fn(repo)
	}
	return total
}

func selectFeatured(repos []gh.Repository) []FeaturedProject {
	if len(repos) == 0 {
		return nil
	}
	const limit = 5
	type scored struct {
		repo  gh.Repository
		score float64
	}
	items := make([]scored, 0, limit)
	sixMonthsAgo := time.Now().AddDate(0, -6, 0)
	for _, repo := range repos {
		score := float64(repo.Stars)*2 + float64(repo.Forks) + float64(repo.Watchers)/2
		if repo.ReadmeStatus == gh.ReadmeStatusVerifiedPresent || repo.HasReadme {
			score += 35
		}
		if repo.ReleaseStatus == gh.ReleaseStatusVerifiedPresent || repo.HasReleases {
			score += 15
		}
		if repo.HasLicense {
			score += 15
		}
		if !repo.Archived {
			score += 20
		}
		if !repo.Fork {
			score += 10
		}
		if repo.UpdatedAt.After(sixMonthsAgo) {
			score += 25
		}
		index := sort.Search(len(items), func(i int) bool { return items[i].score < score })
		if index >= limit {
			continue
		}
		if len(items) < limit {
			items = append(items, scored{})
		}
		copy(items[index+1:], items[index:])
		items[index] = scored{repo: repo, score: score}
	}
	selected := make([]FeaturedProject, 0, len(items))
	for _, item := range items {
		selected = append(selected, FeaturedProject{Repository: item.repo, Why: explainFeatured(item.repo)})
	}
	return selected
}

func explainFeatured(repo gh.Repository) []string {
	var reasons []string
	if repo.Stars > 0 {
		reasons = append(reasons, fmt.Sprintf("%d stars", repo.Stars))
	}
	if repo.UpdatedAt.After(time.Now().AddDate(0, -6, 0)) {
		reasons = append(reasons, "Recent maintenance")
	}
	if repo.ReadmeStatus == gh.ReadmeStatusVerifiedPresent || repo.HasReadme {
		reasons = append(reasons, "Verified README")
	} else if repo.HasDocs {
		reasons = append(reasons, "Documentation metadata")
	}
	if repo.ReleaseStatus == gh.ReleaseStatusVerifiedPresent || repo.HasReleases {
		reasons = append(reasons, "Published releases")
	}
	if repo.HasLicense {
		reasons = append(reasons, "License metadata")
	}
	if repo.Forks > 0 {
		reasons = append(reasons, "Community engagement")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "Public repository footprint")
	}
	return reasons
}

func buildRepositoryLinks(featured []FeaturedProject) []string {
	links := make([]string, 0, len(featured))
	for _, item := range featured {
		if item.Repository.URL != "" {
			links = append(links, fmt.Sprintf("- [%s](%s)", item.Repository.Name, item.Repository.URL))
		}
	}
	return links
}

func buildResumeMarkdown(profile gh.UserProfile, featured []FeaturedProject) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s\n\n", profile.DisplayName))
	b.WriteString(fmt.Sprintf("GitHub: %s\n\n", profile.ProfileURL))
	b.WriteString("## Featured Projects\n")
	for _, project := range featured {
		b.WriteString(fmt.Sprintf("- %s\n  - %s\n  - Language: %s\n  - GitHub: %s\n", project.Repository.Name, project.Repository.Description, project.Repository.Language, project.Repository.URL))
	}
	return b.String()
}

func buildPortfolioMarkdown(profile gh.UserProfile, featured []FeaturedProject, signals []Signal) string {
	var b strings.Builder
	b.WriteString("# Portfolio Summary\n\n")
	b.WriteString(fmt.Sprintf("## %s\n\n", profile.DisplayName))
	b.WriteString("### Featured Projects\n")
	for _, project := range featured {
		b.WriteString(fmt.Sprintf("- %s: %s\n", project.Repository.Name, strings.Join(project.Why, ", ")))
	}
	b.WriteString("\n### Engineering Signals\n")
	for _, signal := range signals {
		b.WriteString(fmt.Sprintf("- %s: %s\n", signal.Name, signal.Level))
	}
	return b.String()
}

func buildReportMarkdown(profile gh.UserProfile, repos []gh.Repository, langs []LanguageBucket, languageBasis string, languageSampleSize int, signals []Signal, featured []FeaturedProject) string {
	var b strings.Builder
	b.WriteString("# GitIntel Report\n\n")
	b.WriteString(fmt.Sprintf("## Profile\n- User: %s\n- Public repositories: %d\n- Followers: %d\n- Following: %d\n\n", profile.Username, len(repos), profile.Followers, profile.Following))
	b.WriteString("## Repository language distribution\n")
	if languageBasis == "code_bytes_sampled" {
		b.WriteString(fmt.Sprintf("Based on code bytes from %d sampled repositories.\n", languageSampleSize))
	} else {
		b.WriteString("Based on repository frequency; this is an estimate, not a code-volume distribution.\n")
	}
	for _, lang := range langs {
		b.WriteString(fmt.Sprintf("- %s: %.1f%%\n", lang.Name, lang.Percentage))
	}
	b.WriteString("\n## Engineering Signals\n")
	for _, signal := range signals {
		b.WriteString(fmt.Sprintf("- %s: %s (%s confidence)\n", signal.Name, signal.Classification, signal.Confidence))
		for _, repository := range signal.Repositories {
			b.WriteString(fmt.Sprintf("  - Supporting repository: %s\n", repository))
		}
		metricNames := make([]string, 0, len(signal.Metric))
		for name := range signal.Metric {
			metricNames = append(metricNames, name)
		}
		sort.Strings(metricNames)
		for _, name := range metricNames {
			b.WriteString(fmt.Sprintf("  - Observed metric, %s: %.0f\n", name, signal.Metric[name]))
		}
		for _, item := range signal.Evidence {
			b.WriteString(fmt.Sprintf("  - Evidence: %s\n", item))
		}
		for _, limitation := range signal.Limitations {
			b.WriteString(fmt.Sprintf("  - Limitation: %s\n", limitation))
		}
	}
	b.WriteString("\n## Overall limitations\n")
	b.WriteString("- Signals rely on collected public repository metadata and timestamps. They do not establish code quality, contribution content, engineering competence, or AI authorship.\n")
	b.WriteString("\n## Featured repositories\n")
	for _, project := range featured {
		b.WriteString(fmt.Sprintf("- %s\n", project.Repository.Name))
	}
	return b.String()
}

func buildProfileReadmeMarkdown(profile gh.UserProfile, featured []FeaturedProject, signals []Signal, langs []LanguageBucket, basis string) string {
	var b strings.Builder
	name := profile.DisplayName
	if name == "" {
		name = profile.Username
	}
	b.WriteString(fmt.Sprintf("# Hi there, I'm %s 👋\n\n", name))
	if profile.Bio != "" {
		b.WriteString(fmt.Sprintf("> %s\n\n", profile.Bio))
	}

	b.WriteString("### ⚡ Engineering Signals & Verified Telemetry\n\n")
	b.WriteString("| Engineering Signal | Level | Observable Evidence |\n")
	b.WriteString("| :--- | :---: | :--- |\n")
	for _, s := range signals {
		evidenceStr := strings.Join(s.Evidence, "; ")
		if evidenceStr == "" {
			evidenceStr = "Verified metadata"
		}
		b.WriteString(fmt.Sprintf("| **%s** | `%s` | %s |\n", s.Name, s.Level, evidenceStr))
	}
	b.WriteString("\n")

	if len(langs) > 0 {
		b.WriteString("### 🛠️ Core Technologies & Codebase Volume\n\n")
		if basis == "code_bytes_sampled" {
			b.WriteString("*Calculated by actual code bytes across top repositories:*\n\n")
		} else {
			b.WriteString("*Calculated across active repositories:*\n\n")
		}
		for _, l := range langs {
			barLength := int(l.Percentage / 5)
			if barLength < 1 && l.Percentage > 0 {
				barLength = 1
			}
			if barLength > 20 {
				barLength = 20
			}
			bar := strings.Repeat("█", barLength) + strings.Repeat("░", 20-barLength)
			b.WriteString(fmt.Sprintf("- **%-14s** `%s` **%.1f%%**\n", l.Name, bar, l.Percentage))
		}
		b.WriteString("\n")
	}

	if len(featured) > 0 {
		b.WriteString("### 🚀 Highlighted Repositories\n\n")
		for _, p := range featured {
			repo := p.Repository
			badges := ""
			if repo.HasReadme {
				badges += " `README`"
			}
			if repo.HasReleases {
				badges += " `Release Tagged`"
			}
			if len(p.Why) > 0 {
				badges += " `★ " + strings.Join(p.Why, "` `★ ") + "`"
			}
			desc := repo.Description
			if desc == "" {
				desc = "No description provided."
			}
			b.WriteString(fmt.Sprintf("- **[%s](%s)** — %s\n", repo.Name, repo.URL, desc))
			b.WriteString(fmt.Sprintf("  - Tech: **%s** • Stars: ⭐ %d • Forks: 🍴 %d%s\n", repo.Language, repo.Stars, repo.Forks, badges))
		}
		b.WriteString("\n")
	}

	b.WriteString("---\n")
	b.WriteString("<sub>⚡ Profile telemetry verified deterministically by [GitIntel](https://github.com/Hosseinghorbani0/GitIntel) • Local-first, evidence-grounded developer intelligence.</sub>\n")
	return b.String()
}

func buildAIPromptMarkdown(profile gh.UserProfile, featured []FeaturedProject, signals []Signal, langs []LanguageBucket) string {
	var b strings.Builder
	b.WriteString("You are an elite executive tech recruiter and resume strategist.\n")
	b.WriteString("Below is 100% verified, observable GitHub engineering telemetry for the developer below, extracted by GitIntel.\n")
	b.WriteString("Treat this telemetry as factual ground truth about the candidate's actual codebase volume and artifacts:\n\n")
	b.WriteString(fmt.Sprintf("Candidate: %s (@%s)\n", profile.DisplayName, profile.Username))
	b.WriteString(fmt.Sprintf("Profile URL: %s\n\n", profile.ProfileURL))

	b.WriteString("=== VERIFIED ENGINEERING SIGNALS ===\n")
	for _, s := range signals {
		b.WriteString(fmt.Sprintf("- %s: %s (Evidence: %s)\n", s.Name, s.Level, strings.Join(s.Evidence, "; ")))
	}
	b.WriteString("\n=== VERIFIED TECH STACK (BY BYTE VOLUME) ===\n")
	for _, l := range langs {
		b.WriteString(fmt.Sprintf("- %s: %.1f%%\n", l.Name, l.Percentage))
	}
	b.WriteString("\n=== KEY VERIFIED REPOSITORIES ===\n")
	for _, p := range featured {
		b.WriteString(fmt.Sprintf("- %s: %s (Language: %s, Signals: %s, URL: %s)\n", p.Repository.Name, p.Repository.Description, p.Repository.Language, strings.Join(p.Why, ", "), p.Repository.URL))
	}
	b.WriteString("\n=== INSTRUCTION ===\n")
	b.WriteString("Write a high-impact, professional resume technical summary and STAR-format project bullet points tailored for Senior Software Engineer applications. Ground all statements strictly in the verified evidence above. Emphasize actual engineering practices: CI/CD automation, testing artifacts, release cadences, and codebase volume. Do not invent unverified achievements.\n")
	return b.String()
}

