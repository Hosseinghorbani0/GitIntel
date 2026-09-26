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
	Activity             ActivityMetrics        `json:"activity"`
	Signals              []Signal               `json:"signals"`
	FeaturedProjects     []FeaturedProject      `json:"featured_projects"`
	RepositoryLinks      []string               `json:"repository_links"`
	ResumeMarkdown       string                 `json:"resume_markdown"`
	PortfolioMarkdown    string                 `json:"portfolio_markdown"`
	ReportMarkdown       string                 `json:"report_markdown"`
}

type Engine struct{}

func NewEngine() *Engine { return &Engine{} }

func (e *Engine) Analyze(profile gh.UserProfile, repos []gh.Repository) Analysis {
	metrics := map[string]interface{}{
		"total_repositories":    len(repos),
		"original_repositories": countOriginal(repos),
		"forked_repositories":   countForked(repos),
		"archived_repositories": countArchived(repos),
		"total_stars":           sum(repos, func(r gh.Repository) int { return r.Stars }),
		"total_forks":           sum(repos, func(r gh.Repository) int { return r.Forks }),
		"total_open_issues":     sum(repos, func(r gh.Repository) int { return r.OpenIssues }),
	}
	langs := buildLanguageDistribution(repos)
	signals := buildSignals(repos)
	featured := selectFeatured(repos)
	report := buildReportMarkdown(profile, repos, langs, signals, featured)
	analysis := Analysis{
		Profile:              profile,
		Metrics:              metrics,
		Summary:              buildSummary(profile, repos, langs),
		LanguageDistribution: langs,
		Activity:             buildActivityMetrics(repos),
		Signals:              signals,
		FeaturedProjects:     featured,
		RepositoryLinks:      buildRepositoryLinks(featured),
		ResumeMarkdown:       buildResumeMarkdown(profile, featured),
		PortfolioMarkdown:    buildPortfolioMarkdown(profile, featured, signals),
		ReportMarkdown:       report,
	}
	return analysis
}

func buildSummary(profile gh.UserProfile, repos []gh.Repository, langs []LanguageBucket) string {
	primary := "General-purpose full-stack activity"
	if len(langs) > 0 {
		primary = langs[0].Name
	}
	if len(repos) == 0 {
		return "No public repositories were returned for this profile, so repository-based engineering signals cannot be assessed."
	}
	maintained := countRecentlyUpdated(repos, 180)
	return fmt.Sprintf("This GitHub profile contains %d public repositories, with the largest repository count in %s. %d repositories have update timestamps within the last 180 days. These metadata signals do not establish code quality, contribution content, or engineering competence.", len(repos), primary, maintained)
}

func buildLanguageDistribution(repos []gh.Repository) []LanguageBucket {
	counts := map[string]int{}
	for _, repo := range repos {
		if strings.TrimSpace(repo.Language) == "" {
			continue
		}
		counts[repo.Language]++
	}
	if len(counts) == 0 {
		return nil
	}
	var buckets []LanguageBucket
	for name, count := range counts {
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
	return Signal{Name: "Code Evolution", Level: level, Classification: level, Confidence: confidence, Weight: 25, Value: float64(updated90), Metric: map[string]float64{"repositories_updated_90d": float64(updated90), "repositories_updated_180d": float64(updated180), "repositories_total": float64(len(repos))}, Evidence: evidence, Limitations: []string{"Uses repository update and push timestamps only; commit history and code changes were not inspected.", "Timestamp recency does not establish contribution volume or code quality."}, Explanation: "Code evolution is inferred from repository update timestamps, not from a definitive measure of technical skill."}
}

func computeMaintenance(repos []gh.Repository) Signal {
	active := countRecentlyUpdated(repos, 180)
	archived := countArchived(repos)
	level := "Limited"
	confidence := "Medium"
	evidence := []string{"Repository maintenance indicators are limited."}
	if active > 0 {
		level = "Strong"
		confidence = "Medium"
		evidence = []string{fmt.Sprintf("%d repositories have recent updates and visible maintenance activity.", active)}
		if archived > 0 {
			level = "Moderate"
			evidence = []string{fmt.Sprintf("%d repositories are archived, while %d repositories remain actively maintained.", archived, active)}
		}
	}
	return Signal{Name: "Maintenance", Level: level, Classification: level, Confidence: confidence, Weight: 20, Value: float64(active), Metric: map[string]float64{"repositories_updated_180d": float64(active), "repositories_archived": float64(archived), "repositories_total": float64(len(repos))}, Evidence: evidence, Limitations: []string{"Uses update timestamps and archived flags only.", "Issue handling, commit contents, and maintenance quality were not inspected."}, Explanation: "Maintenance reflects timestamp recency and archived lifecycle metadata only."}
}

func computeTestingEvidence(repos []gh.Repository) Signal {
	count := 0
	evidence := []string{"Repository-name heuristics found no test/spec markers; repository contents and CI workflows were not inspected."}
	for _, repo := range repos {
		if hasTestingSignals(repo) {
			count++
		}
	}
	if count > 0 {
		evidence = []string{fmt.Sprintf("Repository-name heuristics found test/spec markers in %d repositories; test files and CI workflows were not inspected.", count)}
	}
	level := "Limited"
	confidence := "Medium"
	if count >= 3 {
		level = "Strong"
		confidence = "Medium"
	}
	if count > 0 && count < 3 {
		level = "Moderate"
		confidence = "Medium"
	}
	return Signal{Name: "Testing Evidence", Level: level, Classification: level, Confidence: confidence, Weight: 20, Value: float64(count), Metric: map[string]float64{"repositories_with_test_name_markers": float64(count), "repositories_total": float64(len(repos))}, Evidence: evidence, Limitations: []string{"Uses repository-name heuristics only.", "Test files, test directories, CI workflows, test quality, and correctness were not inspected."}, Explanation: "Repository-name markers are a weak testing hint, not verified test artifacts."}
}

func computeCollaboration(repos []gh.Repository) Signal {
	forkCount := countForked(repos)
	issueCount := sum(repos, func(r gh.Repository) int { return r.OpenIssues })
	level := "Limited"
	confidence := "Low"
	evidence := []string{"No strong collaboration signal available from the public metadata."}
	if forkCount > 0 || issueCount > 0 {
		level = "Moderate"
		confidence = "Medium"
		evidence = []string{fmt.Sprintf("Visible collaboration indicators include %d forked repositories and %d open issues across the public portfolio.", forkCount, issueCount)}
	}
	return Signal{Name: "Collaboration", Level: level, Classification: level, Confidence: confidence, Weight: 15, Value: float64(forkCount), Metric: map[string]float64{"forked_repositories": float64(forkCount), "open_issues": float64(issueCount), "repositories_total": float64(len(repos))}, Evidence: evidence, Limitations: []string{"Uses fork and open-issue counts only.", "Contributors, Pull Requests, discussion quality, and individual behavior were not inspected."}, Explanation: "Fork and open-issue metadata are partial collaboration indicators."}
}

func computeDocumentation(repos []gh.Repository) Signal {
	count := 0
	for _, repo := range repos {
		if repo.HasReadme || repo.HasLicense || repo.HasDocs {
			count++
		}
	}
	level := "Limited"
	confidence := "Medium"
	evidence := []string{"Few repositories show clear documentation patterns."}
	if count > 0 {
		level = "Moderate"
		evidence = []string{fmt.Sprintf("%d repositories show documentation-related metadata hints such as wiki flags, docs markers in repository metadata, or license metadata.", count)}
		if count >= len(repos)/2 {
			level = "Strong"
		}
	}
	return Signal{Name: "Documentation", Level: level, Classification: level, Confidence: confidence, Weight: 10, Value: float64(count), Metric: map[string]float64{"repositories_with_documentation_metadata_hints": float64(count), "repositories_total": float64(len(repos))}, Evidence: evidence, Limitations: []string{"Uses license/wiki and name/description metadata hints only.", "README presence/content and documentation quality were not fetched or inspected."}, Explanation: "Documentation classification is based on repository metadata hints only."}
}

func computeReleaseActivity(repos []gh.Repository) Signal {
	count := countWithReleases(repos)
	level := "Insufficient data"
	confidence := "Low"
	evidence := []string{"Release metadata is not currently fetched by the GitHub client, so release activity is unavailable."}
	if count > 0 {
		level = "Moderate"
		evidence = []string{fmt.Sprintf("%d repositories expose release metadata or release-related signals.", count)}
		if count >= 2 {
			level = "Strong"
		}
	}
	return Signal{Name: "Release / Delivery Signals", Level: level, Classification: level, Confidence: confidence, Weight: 10, Value: float64(count), Metric: map[string]float64{"repositories_with_release_metadata": float64(count), "repositories_total": float64(len(repos))}, Evidence: evidence, Limitations: []string{"The current GitHub client does not query release endpoints.", "No release or deployment conclusion can be made from the current collection."}, Explanation: "Release activity is unavailable unless release metadata is supplied."}
}

func hasTestingSignals(repo gh.Repository) bool {
	name := strings.ToLower(repo.Name)
	if strings.Contains(name, "test") || strings.Contains(name, "spec") || strings.Contains(name, "__tests__") {
		return true
	}
	full := strings.ToLower(repo.FullName)
	if strings.Contains(full, "test") || strings.Contains(full, "spec") {
		return true
	}
	return false
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
		if repo.HasReleases {
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
	type scored struct {
		repo  gh.Repository
		score float64
	}
	var items []scored
	for _, repo := range repos {
		score := float64(repo.Stars)*2 + float64(repo.Forks) + float64(repo.Watchers)/2
		if repo.HasReadme {
			score += 35
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
		if repo.UpdatedAt.After(time.Now().AddDate(0, -6, 0)) {
			score += 25
		}
		items = append(items, scored{repo: repo, score: score})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].score > items[j].score })
	limit := 5
	if len(items) < limit {
		limit = len(items)
	}
	selected := make([]FeaturedProject, 0, limit)
	for _, item := range items[:limit] {
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
	if repo.HasReadme || repo.HasDocs {
		reasons = append(reasons, "Strong documentation")
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

func buildReportMarkdown(profile gh.UserProfile, repos []gh.Repository, langs []LanguageBucket, signals []Signal, featured []FeaturedProject) string {
	var b strings.Builder
	b.WriteString("# GitIntel Report\n\n")
	b.WriteString(fmt.Sprintf("## Profile\n- User: %s\n- Public repositories: %d\n- Followers: %d\n- Following: %d\n\n", profile.Username, len(repos), profile.Followers, profile.Following))
	b.WriteString("## Repository language distribution\n")
	for _, lang := range langs {
		b.WriteString(fmt.Sprintf("- %s: %.1f%%\n", lang.Name, lang.Percentage))
	}
	b.WriteString("\n## Engineering Signals\n")
	for _, signal := range signals {
		b.WriteString(fmt.Sprintf("- %s: %s (%s confidence)\n", signal.Name, signal.Classification, signal.Confidence))
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
