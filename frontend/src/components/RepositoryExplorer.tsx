import { useMemo, useState } from 'react'
import { ArrowDownUp, ArrowUpRight, ChevronDown, Code2, GitFork, LayoutDashboard, ListFilter, Menu, Search, Star } from 'lucide-react'
import type { AnalysisResult, Repository } from '../types'
import type { Messages } from './componentTypes'

interface RepositoryExplorerProps {
  data: AnalysisResult
  analyzedAt: string
  language: 'en' | 'fa'
  t: Messages
  onSelectRepository: (repository: Repository) => void
}

export function RepositoryExplorer({ data, analyzedAt, language, t, onSelectRepository }: RepositoryExplorerProps) {
  const [repoSearch, setRepoSearch] = useState('')
  const [languageFilter, setLanguageFilter] = useState('')
  const [topicFilter, setTopicFilter] = useState('')
  const [licenseFilter, setLicenseFilter] = useState('')
  const [repoType, setRepoType] = useState('all')
  const [archivedFilter, setArchivedFilter] = useState('all')
  const [minimumStars, setMinimumStars] = useState('')
  const [recencyFilter, setRecencyFilter] = useState('all')
  const [repoSort, setRepoSort] = useState('updated')
  const [repoLayout, setRepoLayout] = useState<'list' | 'grid'>('list')

  const repositoryLanguages = useMemo(() => [...new Set(data.repositories.map((repo) => repo.language).filter(Boolean))].sort(), [data.repositories])
  const repositoryTopics = useMemo(() => [...new Set(data.repositories.flatMap((repo) => repo.topics ?? []))].sort(), [data.repositories])
  const repositoryLicenses = useMemo(() => [...new Set(data.repositories.map((repo) => repo.license).filter(Boolean))].sort(), [data.repositories])
  const filteredRepositories = useMemo(() => {
    const searchTerm = repoSearch.trim().toLowerCase()
    const starsThreshold = minimumStars === '' ? null : Number(minimumStars)
    const recencyDays = recencyFilter === 'all' ? null : Number(recencyFilter)
    const analysisTimestamp = Date.parse(analyzedAt)
    const recentBoundary = recencyDays === null || !Number.isFinite(analysisTimestamp)
      ? null
      : analysisTimestamp - recencyDays * 24 * 60 * 60 * 1000
    return [...data.repositories]
      .filter((repo) => !searchTerm || `${repo.name} ${repo.description} ${(repo.topics ?? []).join(' ')}`.toLowerCase().includes(searchTerm))
      .filter((repo) => !languageFilter || repo.language === languageFilter)
      .filter((repo) => !topicFilter || (repo.topics ?? []).includes(topicFilter))
      .filter((repo) => !licenseFilter || repo.license === licenseFilter)
      .filter((repo) => repoType === 'all' || (repoType === 'forks' ? repo.fork : !repo.fork))
      .filter((repo) => archivedFilter === 'all' || (archivedFilter === 'archived' ? repo.archived : !repo.archived))
      .filter((repo) => starsThreshold === null || (Number.isFinite(starsThreshold) && repo.stars >= starsThreshold))
      .filter((repo) => {
        if (recentBoundary === null) return true
        const updatedAt = Date.parse(repo.pushed_at || repo.updated_at)
        return Number.isFinite(updatedAt) && updatedAt >= recentBoundary
      })
      .sort((left, right) => repoSort === 'stars'
        ? right.stars - left.stars
        : repoSort === 'name'
          ? left.name.localeCompare(right.name)
          : Date.parse(right.updated_at) - Date.parse(left.updated_at))
  }, [data.repositories, analyzedAt, repoSearch, languageFilter, topicFilter, licenseFilter, repoType, archivedFilter, minimumStars, recencyFilter, repoSort])

  const activeFilterCount = [
    repoSearch.trim() !== '',
    languageFilter !== '',
    topicFilter !== '',
    licenseFilter !== '',
    repoType !== 'all',
    archivedFilter !== 'all',
    minimumStars !== '',
    recencyFilter !== 'all',
  ].filter(Boolean).length

  const clearFilters = () => {
    setRepoSearch('')
    setLanguageFilter('')
    setTopicFilter('')
    setLicenseFilter('')
    setRepoType('all')
    setArchivedFilter('all')
    setMinimumStars('')
    setRecencyFilter('all')
  }

  const repositoryCard = (repo: Repository) => (
    <article className="repo-card" key={repo.full_name}>
      <div className="repo-card-heading">
        <span className="repo-icon"><Code2 size={17} aria-hidden="true" /></span>
        <button className="repo-name" dir="ltr" aria-haspopup="dialog" onClick={() => onSelectRepository(repo)}>{repo.name}</button>
        <a className="icon-link" href={repo.url} target="_blank" rel="noreferrer" aria-label={`${t.github}: ${repo.name}`}><ArrowUpRight size={16} /></a>
      </div>
      <p className="repo-description"><bdi dir="auto">{repo.description || t.noDescription}</bdi></p>
      <div className="repo-meta">
        <span>{repo.language || t.languageUnknown}</span>
        <span><Star size={13} aria-hidden="true" />{repo.stars}</span>
        <span><GitFork size={13} aria-hidden="true" />{repo.forks}</span>
        <span>{t.updated} <bdi dir="ltr">{new Date(repo.updated_at).toLocaleDateString(language === 'fa' ? 'fa-IR' : 'en-US', { month: 'short', year: 'numeric' })}</bdi></span>
      </div>
      <div className="repo-tags">
        {repo.fork && <span>{t.fork}</span>}{repo.archived && <span>{t.archived}</span>}
        {repo.has_readme && <span>{t.readmeEvidence}</span>}{repo.has_releases && <span>{t.releaseEvidence}</span>}
      </div>
    </article>
  )

  return (
    <>
      <div className="page-title">
        <div><span className="section-kicker">{t.repositoryIndex}</span><h1>{t.repoExplorer}</h1><p>{t.repositoryExplorerIntro}</p></div>
        <span className="result-count" dir="ltr">{filteredRepositories.length} / {data.repositories.length}</span>
      </div>
      <section className="repo-explorer">
        <div className="repo-toolbar">
          <label className="search-control"><Search size={16} /><input type="search" value={repoSearch} onChange={(event) => setRepoSearch(event.target.value)} placeholder={t.searchRepositories} aria-label={t.searchRepositories} /><kbd>/</kbd></label>
          <label className="select-control"><ListFilter size={15} /><select value={languageFilter} onChange={(event) => setLanguageFilter(event.target.value)} aria-label={t.filterLanguage}><option value="">{t.allLanguages}</option>{repositoryLanguages.map((item) => <option key={item} value={item}>{item}</option>)}</select><ChevronDown size={14} /></label>
          <label className="select-control"><select value={repoType} onChange={(event) => setRepoType(event.target.value)} aria-label={t.filterType}><option value="all">{t.allTypes}</option><option value="source">{t.sourceRepositories}</option><option value="forks">{t.forks}</option></select><ChevronDown size={14} /></label>
          <label className="select-control"><ArrowDownUp size={15} /><select value={repoSort} onChange={(event) => setRepoSort(event.target.value)} aria-label={t.sortRepositories}><option value="updated">{t.sortRecentlyUpdated}</option><option value="stars">{t.sortMostStars}</option><option value="name">{t.sortName}</option></select><ChevronDown size={14} /></label>
          <div className="layout-toggle" role="group" aria-label={t.repositoryLayout}><button aria-pressed={repoLayout === 'list'} onClick={() => setRepoLayout('list')} title={t.listView}><Menu size={16} /></button><button aria-pressed={repoLayout === 'grid'} onClick={() => setRepoLayout('grid')} title={t.gridView}><LayoutDashboard size={15} /></button></div>
        </div>
        <div className="repo-filter-grid">
          <label className="select-control"><select value={topicFilter} onChange={(event) => setTopicFilter(event.target.value)} aria-label={t.filterTopic}><option value="">{t.allTopics}</option>{repositoryTopics.map((item) => <option key={item} value={item}>{item}</option>)}</select><ChevronDown size={14} /></label>
          <label className="select-control"><select value={licenseFilter} onChange={(event) => setLicenseFilter(event.target.value)} aria-label={t.filterLicense}><option value="">{t.allLicenses}</option>{repositoryLicenses.map((item) => <option key={item} value={item}>{item}</option>)}</select><ChevronDown size={14} /></label>
          <label className="select-control"><select value={archivedFilter} onChange={(event) => setArchivedFilter(event.target.value)} aria-label={t.filterArchiveStatus}><option value="all">{t.allArchiveStatuses}</option><option value="active">{t.activeRepositories}</option><option value="archived">{t.archivedRepositories}</option></select><ChevronDown size={14} /></label>
          <label className="number-control"><span>{t.minimumStars}</span><input type="number" min="0" step="1" value={minimumStars} onChange={(event) => setMinimumStars(event.target.value)} aria-label={t.minimumStars} /></label>
          <label className="select-control"><span>{t.updatedWithin}</span><select value={recencyFilter} onChange={(event) => setRecencyFilter(event.target.value)} aria-label={t.updatedWithin}><option value="all">{t.anyTime}</option><option value="30">{t.last30Days}</option><option value="90">{t.last90Days}</option><option value="365">{t.lastYear}</option></select><ChevronDown size={14} /></label>
        </div>
        <div className="repo-filter-status" aria-live="polite">
          <span><bdi dir="ltr">{filteredRepositories.length} / {data.repositories.length}</bdi> {t.repositoriesAnalyzed}</span>
          {activeFilterCount > 0 && <button type="button" className="filter-reset" onClick={clearFilters}>{t.clearFilters} · {activeFilterCount}</button>}
        </div>
        {filteredRepositories.length
          ? <div className={`repo-grid ${repoLayout === 'list' ? 'repo-list-layout' : ''}`}>{filteredRepositories.map(repositoryCard)}</div>
          : <div className="empty-inline"><Search size={19} /><strong>{t.noRepositoriesFound}</strong><span>{t.adjustRepositoryFilters}</span></div>}
      </section>
    </>
  )
}
