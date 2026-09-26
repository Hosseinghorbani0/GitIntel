export interface UserProfile {
  username: string
  display_name: string
  avatar_url: string
  bio: string
  location: string
  company: string
  blog: string
  profile_url: string
  created_at: string
  public_repos: number
  followers: number
  following: number
}

export interface Repository {
  name: string
  full_name: string
  description: string
  url: string
  homepage: string
  stars: number
  forks: number
  watchers: number
  open_issues: number
  language: string
  topics: string[]
  created_at: string
  updated_at: string
  pushed_at: string
  default_branch: string
  license: string
  archived: boolean
  fork: boolean
  visibility: string
  has_docs: boolean
  has_readme: boolean
  has_license: boolean
  has_releases: boolean
}

export interface LanguageBucket {
  name: string
  percentage: number
  count: number
}

export interface Signal {
  name: string
  level: string
  classification: string
  confidence: string
  weight: number
  value: number
  metric: Record<string, number>
  evidence: string[]
  limitations: string[]
  explanation: string
}

export interface EvidenceContext {
  schema_version: string
  repository_data_untrusted: boolean
  profile: Record<string, string | number>
  repositories: Record<string, unknown>
  signals: Array<Record<string, unknown>>
  limitations: string[]
}

export interface FeaturedProject {
  repository: Repository
  why: string[]
}

export interface ActivityMetrics {
  commits: string
  pull_requests: string
  issues: string
  releases: string
  repository_updates: string
}

export interface Analysis {
  profile: UserProfile
  metrics: Record<string, number | string | boolean>
  summary: string
  language_distribution: LanguageBucket[]
  activity: ActivityMetrics
  signals: Signal[]
  featured_projects: FeaturedProject[]
  repository_links: string[]
  resume_markdown: string
  portfolio_markdown: string
  report_markdown: string
}

export interface RateLimitInfo {
  limit: number
  remaining: number
  reset: string
}

export interface AnalysisResult {
  profile: UserProfile
  repositories: Repository[]
  analysis: Analysis
  rate_limit: RateLimitInfo
  evidence: EvidenceContext
}

export interface ApiResponse<T> {
  success: boolean
  data: T
  error?: {
    code: string
    message: string
  }
}
