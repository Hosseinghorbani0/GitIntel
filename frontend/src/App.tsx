import { useEffect, useMemo, useState, type FormEvent } from 'react'
import {
  Activity, ArrowDownUp, ArrowUpRight, BookOpen, BriefcaseBusiness, Check, ChevronDown,
  CircleAlert, Clock3, Code2, Download, FileText, GitBranch, GitFork, Globe2, LayoutDashboard,
  ListFilter, LoaderCircle, LockKeyhole, Menu, Network, Search, Settings2, ShieldCheck,
  Star, X,
} from 'lucide-react'
import './App.css'
import { getLLMStatus, interpretAnalysis, testLLMModel, type InterpretationResult, type LLMConfigStatus } from './llm'
import { messages, type Language } from './i18n'
import { analyzeProfile, copyText } from './services/api'
import type { AnalysisResult, Repository, Signal } from './types'

type ViewKey = 'analyze' | 'dashboard' | 'repositories' | 'repository' | 'signals' | 'timeline' | 'reports' | 'portfolio' | 'settings'

function App() {
  const [language, setLanguage] = useState<Language>(() => window.localStorage.getItem('gitintel-language') === 'fa' ? 'fa' : 'en')
  const t = messages[language]
  const [view, setView] = useState<ViewKey>('analyze')
  const [mobileMore, setMobileMore] = useState(false)
  const [username, setUsername] = useState('Hosseinghorbani0')
  const [token, setToken] = useState('')
  const [loading, setLoading] = useState(false)
  const [llmLoading, setLLMLoading] = useState(false)
  const [interpretationLoading, setInterpretationLoading] = useState(false)
  const [interpretation, setInterpretation] = useState<InterpretationResult | null>(null)
  const [interpretationError, setInterpretationError] = useState('')
  const [llmResult, setLLMResult] = useState<{ provider: string; model: string; status: string; latency_ms: number; tokens?: number; response: string; error?: string } | null>(null)
  const [llmConfig, setLLMConfig] = useState<LLMConfigStatus>({ provider: 'huggingface', model: 'Qwen/Qwen3.8-2.4T-A95B:novita', credential_status: 'not configured', provider_status: 'unverified', analysis_reports_enabled: false })
  const [llmError, setLLMError] = useState('')
  const [error, setError] = useState('')
  const [toast, setToast] = useState('')
  const [data, setData] = useState<AnalysisResult | null>(null)
  const [analyzedAt, setAnalyzedAt] = useState('')
  const [repoSearch, setRepoSearch] = useState('')
  const [languageFilter, setLanguageFilter] = useState('')
  const [repoType, setRepoType] = useState('all')
  const [repoSort, setRepoSort] = useState('updated')
  const [repoLayout, setRepoLayout] = useState<'list' | 'grid'>('list')
  const [selectedRepository, setSelectedRepository] = useState<Repository | null>(null)
  const [portfolioMode, setPortfolioMode] = useState<'resume' | 'portfolio'>('resume')

  useEffect(() => {
    document.documentElement.lang = language
    document.documentElement.dir = language === 'fa' ? 'rtl' : 'ltr'
    window.localStorage.setItem('gitintel-language', language)
  }, [language])

  useEffect(() => {
    void getLLMStatus().then(setLLMConfig).catch(() => undefined)
  }, [])

  useEffect(() => {
    if (!toast) return
    const timeout = window.setTimeout(() => setToast(''), 2600)
    return () => window.clearTimeout(timeout)
  }, [toast])

  const summary = data?.analysis.summary
  const repositoryLanguages = useMemo(() => [...new Set((data?.repositories ?? []).map((repo) => repo.language).filter(Boolean))].sort(), [data])
  const filteredRepositories = useMemo(() => {
    const searchTerm = repoSearch.trim().toLowerCase()
    return [...(data?.repositories ?? [])]
      .filter((repo) => !searchTerm || `${repo.name} ${repo.description} ${repo.topics.join(' ')}`.toLowerCase().includes(searchTerm))
      .filter((repo) => !languageFilter || repo.language === languageFilter)
      .filter((repo) => repoType === 'all' || (repoType === 'forks' ? repo.fork : !repo.fork))
      .sort((left, right) => repoSort === 'stars'
        ? right.stars - left.stars
        : repoSort === 'name'
          ? left.name.localeCompare(right.name)
          : Date.parse(right.updated_at) - Date.parse(left.updated_at))
  }, [data, repoSearch, languageFilter, repoType, repoSort])

  const handleAnalyze = async (event?: FormEvent) => {
    event?.preventDefault()
    setLoading(true)
    setError('')
    try {
      const result = await analyzeProfile(username, token)
      setData(result)
      setAnalyzedAt(new Date().toISOString())
      setInterpretation(null)
      setInterpretationError('')
      setSelectedRepository(null)
      setView('dashboard')
      setToast(t.analysisComplete)
    } catch (err) {
      setError(err instanceof Error ? err.message : t.analysisFailed)
    } finally {
      setLoading(false)
    }
  }

  const handleInterpret = async () => {
    if (!data) return
    setInterpretationLoading(true)
    setInterpretationError('')
    try {
      setInterpretation(await interpretAnalysis(data.profile.username, language))
    } catch (err) {
      setInterpretation(null)
      setInterpretationError(err instanceof Error ? err.message : 'LLM_PROVIDER_ERROR')
    } finally {
      setInterpretationLoading(false)
    }
  }

  const interpretationErrorMessage = interpretationError === 'LLM_ANALYSIS_DISABLED'
    ? t.interpretationDisabled
    : interpretationError === 'LLM_PAYMENT_REQUIRED'
      ? t.paymentRequired
      : interpretationError === 'LLM_RATE_LIMITED'
        ? t.rateLimited
        : interpretationError === 'LLM_INVALID_CREDENTIAL'
          ? t.invalidCredential
          : interpretationError === 'LLM_TIMEOUT'
            ? t.timeoutError
            : interpretationError === 'ANALYSIS_NOT_FOUND'
              ? t.noCachedAnalysis
              : interpretationError === 'INVALID_LLM_CONFIG'
                ? t.invalidProviderConfig
                : interpretationError ? t.aiUnavailable : ''

  const handleLLMTest = async () => {
    setLLMLoading(true)
    setLLMError('')
    try {
      const result = await testLLMModel(language)
      setLLMResult(result)
      setLLMConfig((current) => ({ ...current, credential_status: result.credential_status, provider_status: 'available' }))
    } catch (err) {
      const message = err instanceof Error ? err.message : ''
      if (message === 'INVALID_CREDENTIAL') setLLMConfig((current) => ({ ...current, credential_status: 'invalid' }))
      else if (message === 'RATE_LIMITED') setLLMConfig((current) => ({ ...current, credential_status: 'temporarily unavailable' }))
      else if (message === 'PAYMENT_REQUIRED') setLLMConfig((current) => ({ ...current, provider_status: 'payment_required' }))
      setLLMError(message === 'RATE_LIMITED' ? t.rateLimited : message === 'PAYMENT_REQUIRED' ? t.paymentRequired : message === 'INVALID_CREDENTIAL' ? t.invalidCredential : message || t.modelFailure)
    } finally {
      setLLMLoading(false)
    }
  }

  const credentialLabel = llmConfig.credential_status === 'active' ? t.active
    : llmConfig.credential_status === 'configured' ? t.configured
      : llmConfig.credential_status === 'invalid' ? t.invalid
        : llmConfig.credential_status === 'temporarily unavailable' ? t.temporarilyUnavailable : t.notConfigured
  const statusLabel = llmConfig.provider_status === 'payment_required' ? t.providerPaymentRequired
    : llmConfig.provider_status === 'rate_limited' ? t.rateLimited
      : !llmConfig.analysis_reports_enabled ? t.aiDisabled
        : llmResult?.status === 'Success' ? t.success
          : llmResult?.status === 'Rate Limited' ? t.rateLimited
            : llmResult?.status === 'Testing' ? t.testingStatus
              : llmResult ? t.errorStatus : t.ready

  const copyValue = async (value: string) => {
    try {
      await copyText(value)
      setToast(t.copied)
    } catch {
      setToast(t.copyUnavailable)
    }
  }

  const downloadFile = (filename: string, content: string, type: string) => {
    const url = URL.createObjectURL(new Blob([content], { type }))
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    link.click()
    URL.revokeObjectURL(url)
  }

  const navItems = [
    { id: 'analyze' as const, label: t.nav.analyze, icon: Search },
    { id: 'dashboard' as const, label: t.nav.dashboard, icon: LayoutDashboard },
    { id: 'repositories' as const, label: t.nav.repositories, icon: Code2 },
    { id: 'signals' as const, label: t.nav.signals, icon: Activity },
    { id: 'timeline' as const, label: t.nav.timeline, icon: Clock3 },
    { id: 'reports' as const, label: t.nav.reports, icon: FileText },
    { id: 'portfolio' as const, label: t.nav.portfolio, icon: BriefcaseBusiness },
    { id: 'settings' as const, label: t.nav.settings, icon: Settings2 },
  ]
  const activeLabel = view === 'repository' ? t.repositoryDetail : navItems.find((item) => item.id === view)?.label ?? t.nav.dashboard

  const renderSignal = (signal: Signal, index: number) => {
    const level = signal.level.toLowerCase()
    const stateClass = level.includes('insufficient') ? 'is-muted' : level.includes('limited') ? 'is-caution' : 'is-positive'
    return (
      <article key={signal.name} className={`signal-card ${stateClass}`}>
        <div className="signal-card-top">
          <span className="signal-index">0{index + 1}</span>
          <span className="signal-state">{signal.level}</span>
        </div>
        <h3>{signal.name}</h3>
        <p className="signal-explanation">{signal.explanation}</p>
        <div className="signal-evidence-count"><span>{t.evidenceItems}</span><strong>{signal.evidence.length}</strong></div>
        {signal.evidence.length > 0 ? (
          <ul className="evidence-list">{signal.evidence.slice(0, 3).map((item) => <li key={item}><bdi dir="auto">{item}</bdi></li>)}</ul>
        ) : <p className="quiet-empty">{t.insufficientEvidence}</p>}
        {signal.limitations.length > 0 && <details className="signal-limitations"><summary>{t.viewLimitations}</summary><p>{signal.limitations.join(' ')}</p></details>}
      </article>
    )
  }

  const repositoryCard = (repo: Repository) => (
    <article className="repo-card" key={repo.full_name}>
      <div className="repo-card-heading">
        <span className="repo-icon"><Code2 size={17} aria-hidden="true" /></span>
        <button className="repo-name" dir="ltr" onClick={() => { setSelectedRepository(repo); setView('repository') }}>{repo.name}</button>
        <a className="icon-link" href={repo.url} target="_blank" rel="noreferrer" aria-label={`${t.github}: ${repo.name}`}><ArrowUpRight size={16} /></a>
      </div>
      <p className="repo-description">{repo.description || t.noDescription}</p>
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

  const overviewHeader = data && (
    <section className="identity-strip">
      {data.profile.avatar_url ? <img className="avatar" src={data.profile.avatar_url} alt="" /> : <div className="avatar avatar-fallback"><GitBranch size={23} /></div>}
      <div className="identity-copy">
        <div className="identity-name-row"><h1>{data.profile.display_name || data.profile.username}</h1><span className="status-tag"><span />{t.analysisReady}</span></div>
        <div className="identity-meta"><bdi dir="ltr">@{data.profile.username}</bdi><span>·</span><span>{t.analyzedAt} <bdi dir="ltr">{analyzedAt ? new Date(analyzedAt).toLocaleString(language === 'fa' ? 'fa-IR' : 'en-US', { dateStyle: 'medium', timeStyle: 'short' }) : t.unavailable}</bdi></span></div>
        {data.profile.bio && <p>{data.profile.bio}</p>}
      </div>
      <div className="identity-actions"><a className="quiet-button" href={data.profile.profile_url} target="_blank" rel="noreferrer"><GitBranch size={15} />{t.viewGithub}<ArrowUpRight size={13} /></a><button className="quiet-button" onClick={() => { setView('analyze'); window.scrollTo({ top: 0, behavior: 'smooth' }) }}><Search size={15} />{t.newAnalysis}</button></div>
    </section>
  )

  const emptyState = (title: string, description: string) => (
    <div className="empty-state"><div className="empty-icon"><Network size={22} /></div><h2>{title}</h2><p>{description}</p><button className="button-primary" onClick={() => setView('analyze')}>{t.startAnalysis}</button></div>
  )

  const renderLLMSettings = () => (
    <section className="content-section llm-settings">
      <div className="section-heading"><div><span className="section-kicker">{t.optionalLayer}</span><h2>{t.llmTitle}</h2><p>{t.llmDescription}</p></div><span className={`status-tag ${llmConfig.analysis_reports_enabled ? 'status-available' : 'status-unavailable'}`}><span />{statusLabel}</span></div>
      <div className="settings-grid">
        <div className="setting-row"><span>{t.provider}</span><strong>{llmConfig.provider === 'huggingface' ? 'Hugging Face' : llmConfig.provider}</strong></div>
        <div className="setting-row"><span>{t.model}</span><strong className="ltr-value">{llmConfig.model}</strong></div>
        <div className="setting-row"><span>{t.credential}</span><strong>{credentialLabel}</strong></div>
        <div className="setting-row"><span>{t.status}</span><strong>{statusLabel}</strong></div>
        <div className="setting-row"><span>{t.latency}</span><strong className="ltr-value">{llmResult ? `${llmResult.latency_ms}ms` : t.notTested}</strong></div>
        <div className="setting-row"><span>{t.tokens}</span><strong className="ltr-value">{llmResult?.tokens ?? t.notAvailable}</strong></div>
      </div>
      <div className="integrity-note"><ShieldCheck size={17} /><p>{t.llmIntegrity}</p></div>
      <button className="button-secondary" onClick={handleLLMTest} disabled={llmLoading || llmConfig.provider_status === 'payment_required'}>{llmLoading ? <LoaderCircle className="spin" size={16} /> : <Activity size={16} />}{llmLoading ? t.testing : t.testModel}</button>
      {llmError && <div className="inline-error" role="alert"><CircleAlert size={16} />{llmError}</div>}
      {llmResult && <div className="model-output" dir="auto"><span>{t.modelTestResponse}</span><p>{llmResult.response || llmResult.error || t.noResponse}</p></div>}
    </section>
  )

  return (
    <div className="app-shell">
      <aside className="sidebar" aria-label={t.primaryNavigation}>
        <button className="brand-lockup" onClick={() => setView('analyze')} aria-label="GitIntel home"><span className="brand-symbol"><span /><span /><span /></span><span className="brand-word">GitIntel<span className="brand-period">.</span></span></button>
        <div className="workspace-label">{t.workspace}</div>
        <nav className="side-nav">
          {navItems.map(({ id, label, icon: Icon }) => <button key={id} className={`nav-link ${view === id || (view === 'repository' && id === 'repositories') ? 'active' : ''}`} aria-current={view === id || (view === 'repository' && id === 'repositories') ? 'page' : undefined} onClick={() => { setView(id); setMobileMore(false) }}><Icon size={17} strokeWidth={1.8} /><span>{label}</span>{id === 'dashboard' && data && <span className="nav-count">{data.analysis.signals.length}</span>}</button>)}
        </nav>
        <div className="sidebar-spacer" />
        <div className="connection-state"><span className="connection-dot" /><span>{t.githubPublicData}</span></div>
        <div className="sidebar-language"><Globe2 size={15} /><span>{t.language}</span><div className="language-switch"><button type="button" lang="en" aria-pressed={language === 'en'} onClick={() => setLanguage('en')}>EN</button><button type="button" lang="fa" aria-pressed={language === 'fa'} onClick={() => setLanguage('fa')}>فا</button></div></div>
        <div className="sidebar-user"><div className="user-avatar">GI</div><div><strong>{data?.profile.username || 'GitIntel'}</strong><span>{data ? t.localAnalysis : t.personalWorkspace}</span></div></div>
      </aside>

      <div className="workspace-main">
        <header className="mobile-topbar"><button className="mobile-brand" onClick={() => setView('analyze')}><span className="brand-symbol"><span /><span /><span /></span><b>GitIntel<span className="brand-period">.</span></b></button><div className="language-switch"><button type="button" lang="en" aria-pressed={language === 'en'} onClick={() => setLanguage('en')}>EN</button><button type="button" lang="fa" aria-pressed={language === 'fa'} onClick={() => setLanguage('fa')}>فا</button></div></header>
        <header className="workspace-topbar"><div className="breadcrumb"><span>{t.workspace}</span><span>/</span><strong>{activeLabel}</strong></div><div className="topbar-context">{data ? <><span className="context-dot" /><bdi dir="ltr">{data.profile.username}</bdi><span className="context-divider" /><span>{data.repositories.length} {t.repositoriesAnalyzed}</span></> : <span>{t.evidenceWorkspace}</span>}</div></header>
        <main className="main-content">
          {error && <div className="global-error" role="alert"><div className="error-icon"><CircleAlert size={18} /></div><div><strong>{t.analysisFailed}</strong><p>{error}</p><button onClick={() => setError('')}>{t.dismiss}</button></div></div>}
          {view === 'analyze' && <>
            {!data ? <section className="analyze-view">
              <div className="analyze-intro"><span className="section-kicker"><span className="kicker-mark" />{t.evidenceWorkspace}</span><h1>{t.heroTitle}</h1><p>{t.heroText}</p></div>
              <form className="analysis-form" onSubmit={(event) => void handleAnalyze(event)}>
                <label className="field"><span>{t.username}</span><div className="input-with-icon"><GitBranch size={18} /><input dir="ltr" autoComplete="username" value={username} onChange={(event) => setUsername(event.target.value)} placeholder={t.usernamePlaceholder} /><span className="input-prefix">github.com /</span></div></label>
                <label className="field token-field"><span>{t.optionalToken}</span><div className="input-with-icon"><LockKeyhole size={17} /><input dir="ltr" type="password" autoComplete="off" value={token} onChange={(event) => setToken(event.target.value)} placeholder={t.tokenPlaceholder} /></div><small>{t.tokenHelper}</small></label>
                <button className="button-primary analyze-button" type="submit" disabled={loading || !username.trim()}>{loading ? <LoaderCircle className="spin" size={17} /> : <Search size={17} />}{loading ? t.analyzing : t.analyze}<ArrowUpRight size={15} /></button>
              </form>
              <div className="analysis-integrity"><ShieldCheck size={17} /><p>{t.security}</p></div>
              {loading && <div className="loading-panel" role="status" aria-live="polite"><div className="loading-heading"><LoaderCircle className="spin" size={19} /><div><strong>{t.analysisInProgress}</strong><span>{t.analysisProgressDetail}</span></div></div><div className="loading-track"><span /></div><div className="loading-foot"><span>{t.githubRequest}</span><span>{t.awaitingResponse}</span></div></div>}
              <div className="principles-row">{t.features.map((feature, index) => <div className="principle" key={feature}><span>0{index + 1}</span><strong>{feature}</strong></div>)}</div>
              <div className="ai-tag-row" aria-label="AI interpretation layer">
                <span className="ai-tag">AI</span>
                <span>{t.aiSeparateLayer}</span>
              </div>
            </section> : <>
              {overviewHeader}
              <div className="section-title-row"><div><span className="section-kicker">{t.workspace}</span><h1>{t.newAnalysis}</h1></div><p>{t.heroText}</p></div>
              <form className="analysis-form compact-form" onSubmit={(event) => void handleAnalyze(event)}><label className="field"><span>{t.username}</span><div className="input-with-icon"><GitBranch size={17} /><input dir="ltr" value={username} onChange={(event) => setUsername(event.target.value)} /></div></label><button className="button-primary" type="submit" disabled={loading || !username.trim()}>{loading ? <LoaderCircle className="spin" size={16} /> : <Search size={16} />}{loading ? t.analyzing : t.analyze}</button></form>
              {loading && <div className="loading-panel" role="status"><div className="loading-heading"><LoaderCircle className="spin" size={19} /><div><strong>{t.analysisInProgress}</strong><span>{t.analysisProgressDetail}</span></div></div><div className="loading-track"><span /></div></div>}
            </>}
          </>}

          {view !== 'analyze' && !data && emptyState(t.noAnalysisTitle, t.noAnalysisDescription)}

          {data && view === 'dashboard' && <div className="view-stack">
            {overviewHeader}
            <div className="overview-heading"><div><span className="section-kicker">{t.observedOverview}</span><h1>{t.developerProfile}</h1></div><p>{t.deterministicAvailable}</p></div>
            <section className="overview-grid">
              <article className="overview-summary"><div className="summary-label"><span className="summary-mark"><Activity size={16} /></span>{t.evidenceSummary}<span className="deterministic-badge">{t.deterministic}</span></div><p>{summary || t.insufficientEvidence}</p><div className="summary-foot"><span><ShieldCheck size={14} />{t.evidenceOnly}</span><button onClick={() => setView('reports')}>{t.openReport}<ArrowUpRight size={14} /></button></div></article>
              <div className="stat-stack">
                <div className="stat-cell"><span>{t.repositoriesAnalyzed}</span><strong>{data.repositories.length}</strong><small>{t.publicRepositoriesCount} <bdi dir="ltr">{data.profile.public_repos}</bdi></small></div>
                <div className="stat-cell"><span>{t.languagesObserved}</span><strong>{data.analysis.language_distribution?.length ?? 0}</strong><small>{t.metadataOnly}</small></div>
                <div className="stat-cell"><span>{t.profileSince}</span><strong>{new Date(data.profile.created_at).getFullYear()}</strong><small>{t.githubAccount}</small></div>
              </div>
            </section>
            <section className="content-section"><div className="section-heading"><div><span className="section-kicker">{t.observedSignals}</span><h2>{t.engineeringSignals}</h2><p>{t.signalsIntro}</p></div><button className="text-link" onClick={() => setView('signals')}>{t.allSignals}<ArrowUpRight size={14} /></button></div><div className="signal-grid">{data.analysis.signals.slice(0, 3).map(renderSignal)}</div><div className="signals-footer"><span>{t.noDeveloperScore}</span><button onClick={() => setView('signals')}>{t.viewAllSignals}<ArrowUpRight size={14} /></button></div></section>
            <section className="content-section compact-section"><div className="section-heading"><div><span className="section-kicker">{t.projects}</span><h2>{t.featuredProjects}</h2></div><button className="text-link" onClick={() => setView('repositories')}>{t.browseRepositories}<ArrowUpRight size={14} /></button></div>{data.analysis.featured_projects.length ? <div className="repo-grid featured-grid">{data.analysis.featured_projects.slice(0, 3).map((project) => <article className="featured-row" key={project.repository.full_name}><div className="featured-main"><span className="repo-icon"><Code2 size={16} /></span><div><strong dir="ltr">{project.repository.name}</strong><p>{project.repository.description || t.noDescription}</p></div></div><div className="repo-meta"><span>{project.repository.language || t.languageUnknown}</span><span><Star size={13} />{project.repository.stars}</span></div>{project.why.length > 0 && <small>{project.why[0]}</small>}</article>)}</div> : <p className="quiet-empty">{t.noFeaturedProjects}</p>}</section>
          </div>}

          {data && view === 'repositories' && <div className="view-stack">{overviewHeader}<div className="page-title"><div><span className="section-kicker">{t.repositoryIndex}</span><h1>{t.repoExplorer}</h1><p>{t.repositoryExplorerIntro}</p></div><span className="result-count">{filteredRepositories.length} / {data.repositories.length}</span></div><section className="repo-explorer"><div className="repo-toolbar"><label className="search-control"><Search size={16} /><input value={repoSearch} onChange={(event) => setRepoSearch(event.target.value)} placeholder={t.searchRepositories} aria-label={t.searchRepositories} /><kbd>/</kbd></label><label className="select-control"><ListFilter size={15} /><select value={languageFilter} onChange={(event) => setLanguageFilter(event.target.value)} aria-label={t.filterLanguage}><option value="">{t.allLanguages}</option>{repositoryLanguages.map((item) => <option key={item} value={item}>{item}</option>)}</select><ChevronDown size={14} /></label><label className="select-control"><select value={repoType} onChange={(event) => setRepoType(event.target.value)} aria-label={t.filterType}><option value="all">{t.allTypes}</option><option value="source">{t.sourceRepositories}</option><option value="forks">{t.forks}</option></select><ChevronDown size={14} /></label><label className="select-control"><ArrowDownUp size={15} /><select value={repoSort} onChange={(event) => setRepoSort(event.target.value)} aria-label={t.sortRepositories}><option value="updated">{t.sortRecentlyUpdated}</option><option value="stars">{t.sortMostStars}</option><option value="name">{t.sortName}</option></select><ChevronDown size={14} /></label><div className="layout-toggle" role="group" aria-label={t.repositoryLayout}><button aria-pressed={repoLayout === 'list'} onClick={() => setRepoLayout('list')} title={t.listView}><Menu size={16} /></button><button aria-pressed={repoLayout === 'grid'} onClick={() => setRepoLayout('grid')} title={t.gridView}><LayoutDashboard size={15} /></button></div></div>{filteredRepositories.length ? <div className={`repo-grid ${repoLayout === 'list' ? 'repo-list-layout' : ''}`}>{filteredRepositories.map(repositoryCard)}</div> : <div className="empty-inline"><Search size={19} /><strong>{t.noRepositoriesFound}</strong><span>{t.adjustRepositoryFilters}</span></div>}</section></div>}

          {data && view === 'repository' && selectedRepository && <div className="view-stack">{overviewHeader}<button className="back-link" onClick={() => setView('repositories')}>← {t.backToRepositories}</button><section className="repo-detail-head"><div className="repo-icon large"><Code2 size={21} /></div><div className="repo-detail-title"><div><span className="section-kicker">{t.repositoryInspection}</span><h1 dir="ltr">{selectedRepository.full_name}</h1></div><a className="quiet-button" href={selectedRepository.url} target="_blank" rel="noreferrer">{t.viewGithub}<ArrowUpRight size={14} /></a><p>{selectedRepository.description || t.noDescription}</p><div className="repo-meta"><span>{selectedRepository.language || t.languageUnknown}</span><span><Star size={13} />{selectedRepository.stars}</span><span><GitFork size={13} />{selectedRepository.forks}</span><span>{selectedRepository.visibility}</span><span>{selectedRepository.license || t.licenseUnknown}</span></div></div></section><div className="detail-grid">{[{ title: t.activitySection, text: `${t.created} ${new Date(selectedRepository.created_at).toLocaleDateString(language === 'fa' ? 'fa-IR' : 'en-US')} · ${t.updated} ${new Date(selectedRepository.updated_at).toLocaleDateString(language === 'fa' ? 'fa-IR' : 'en-US')}`, icon: Clock3 }, { title: t.testingSection, text: t.testingNotCollected, icon: Check }, { title: t.documentationSection, text: selectedRepository.has_readme || selectedRepository.has_docs ? t.documentationObserved : t.documentationNotObserved, icon: BookOpen }, { title: t.collaborationSection, text: t.collaborationNotCollected, icon: Network }, { title: t.deliverySection, text: selectedRepository.has_releases ? t.releaseMetadataObserved : t.releaseMetadataNotObserved, icon: ArrowUpRight }, { title: t.limitations, text: t.repositoryLimitations, icon: ShieldCheck }].map(({ title, text, icon: Icon }) => <article className="detail-panel" key={title}><div className="detail-panel-title"><Icon size={16} /><h2>{title}</h2></div><p>{text}</p></article>)}</div></div>}

          {data && view === 'signals' && <div className="view-stack">{overviewHeader}<div className="page-title"><div><span className="section-kicker">{t.observableEvidence}</span><h1>{t.engineeringSignals}</h1><p>{t.signalsIntro}</p></div><span className="deterministic-badge">{t.noDeveloperScore}</span></div><div className="signal-grid signal-grid-all">{data.analysis.signals.map(renderSignal)}</div><section className="content-section attention-section"><div className="section-heading"><div><span className="section-kicker">{t.inspectWithContext}</span><h2>{t.attentionSignals}</h2><p>{t.attentionIntro}</p></div></div>{data.analysis.signals.filter((signal) => /limited|insufficient/i.test(signal.level)).length ? <div className="attention-list">{data.analysis.signals.filter((signal) => /limited|insufficient/i.test(signal.level)).map((signal) => <article className="attention-row" key={signal.name}><span className="attention-icon"><CircleAlert size={16} /></span><div><strong>{signal.name}</strong><p>{signal.explanation}</p><small>{signal.limitations.join(' ') || t.cannotConclude}</small></div><span className="signal-state">{signal.level}</span></article>)}</div> : <div className="empty-inline"><Check size={18} /><strong>{t.noAttentionSignals}</strong><span>{t.noAttentionDescription}</span></div>}</section></div>}

          {data && view === 'timeline' && <div className="view-stack">{overviewHeader}<div className="page-title"><div><span className="section-kicker">{t.observedHistory}</span><h1>{t.timeline}</h1><p>{t.timelineIntro}</p></div></div>{data.repositories.length ? <><div className="timeline-list">{data.repositories.flatMap((repo) => [{ key: `${repo.full_name}-created`, date: repo.created_at, title: t.repositoryCreated, repo }, { key: `${repo.full_name}-updated`, date: repo.updated_at, title: t.repositoryUpdated, repo }]).sort((left, right) => Date.parse(right.date) - Date.parse(left.date)).slice(0, 40).map((event) => <article className="timeline-event" key={event.key}><span className="timeline-marker" /><time dateTime={event.date}>{new Date(event.date).toLocaleDateString(language === 'fa' ? 'fa-IR' : 'en-US', { dateStyle: 'medium' })}</time><div><strong>{event.title}</strong><button dir="ltr" onClick={() => { setSelectedRepository(event.repo); setView('repository') }}>{event.repo.full_name}<ArrowUpRight size={13} /></button></div></article>)}</div><p className="timeline-note"><CircleAlert size={15} />{t.timelineLimitation}</p></> : <div className="empty-inline"><Clock3 size={18} /><strong>{t.noTimelineData}</strong><span>{t.noTimelineDescription}</span></div>}</div>}

          {data && view === 'reports' && <div className="view-stack report-view">{overviewHeader}<div className="page-title"><div><span className="section-kicker">{t.deterministic}</span><h1>{t.analysisTitle}</h1><p>{t.reportIntro}</p></div><div className="report-actions"><button className="button-secondary" onClick={() => window.print()}><FileText size={15} />{t.printPDF}</button><button className="button-secondary" onClick={() => downloadFile(`${data.profile.username}-report.md`, data.analysis.report_markdown, 'text/markdown')}><Download size={15} />{t.downloadMarkdown}</button></div></div><article className="report-document"><div className="report-document-head"><div><span className="section-kicker">GITINTEL / ENGINEERING EVIDENCE</span><h2>{data.profile.display_name || data.profile.username}</h2><p><bdi dir="ltr">@{data.profile.username}</bdi> · {t.analyzedAt} <bdi dir="ltr">{analyzedAt ? new Date(analyzedAt).toLocaleDateString(language === 'fa' ? 'fa-IR' : 'en-US') : t.unavailable}</bdi></p></div><span className="deterministic-badge">{t.deterministic}</span></div><pre className="report-markdown" dir="ltr">{data.analysis.report_markdown}</pre><div className="report-method"><ShieldCheck size={16} /><p>{t.reportMethodology}</p></div></article><section className="ai-availability"><div className="ai-label"><span className="ai-indicator" /><div><strong>{t.aiInterpretation}</strong><p>{t.aiSeparateLayer}</p></div></div><span className="status-tag status-unavailable"><span />{statusLabel}</span><p className="ai-explanation">{interpretationErrorMessage || (!llmConfig.analysis_reports_enabled ? (llmConfig.provider_status === 'payment_required' ? t.paymentRequired : t.interpretationDisabled) : t.aiAvailable)}</p><button className="button-secondary" onClick={handleInterpret} disabled={interpretationLoading || !llmConfig.analysis_reports_enabled}>{interpretationLoading ? <LoaderCircle className="spin" size={15} /> : <Activity size={15} />}{interpretationLoading ? t.generatingInterpretation : t.generateInterpretation}</button>{interpretation && <div className="interpretation-output" lang={interpretation.language} dir={interpretation.language === 'fa' ? 'rtl' : 'ltr'}><div className="interpretation-meta" dir="ltr">{interpretation.provider} / {interpretation.model} / {interpretation.latency_ms}ms</div><p>{interpretation.report}</p></div>}</section><div className="report-actions bottom-actions"><button className="button-secondary" onClick={() => void copyValue(data.analysis.report_markdown)}><FileText size={15} />{t.copyReport}</button><button className="button-secondary" onClick={() => downloadFile(`${data.profile.username}-report.json`, JSON.stringify(data, null, 2), 'application/json')}><Download size={15} />{t.downloadJSON}</button></div></div>}

          {data && view === 'portfolio' && <div className="view-stack">{overviewHeader}<div className="page-title"><div><span className="section-kicker">{t.evidenceBasedOutput}</span><h1>{t.nav.portfolio}</h1><p>{t.portfolioIntro}</p></div></div><div className="segmented-control" role="tablist" aria-label={t.portfolioOutput}><button role="tab" aria-selected={portfolioMode === 'resume'} onClick={() => setPortfolioMode('resume')}><BriefcaseBusiness size={15} />{t.resume}</button><button role="tab" aria-selected={portfolioMode === 'portfolio'} onClick={() => setPortfolioMode('portfolio')}><BookOpen size={15} />{t.portfolio}</button></div><section className="portfolio-document"><div className="portfolio-document-head"><div><span className="section-kicker">{t.observedOnGithub}</span><h2>{portfolioMode === 'resume' ? t.resume : t.portfolio}</h2></div><div className="report-actions"><button className="button-secondary" onClick={() => void copyValue(portfolioMode === 'resume' ? data.analysis.resume_markdown : data.analysis.portfolio_markdown)}><FileText size={15} />{t.copyText}</button><button className="button-secondary" onClick={() => downloadFile(`${data.profile.username}-${portfolioMode}.md`, portfolioMode === 'resume' ? data.analysis.resume_markdown : data.analysis.portfolio_markdown, 'text/markdown')}><Download size={15} />{t.downloadMarkdown}</button></div></div><pre className="report-markdown" dir="ltr">{portfolioMode === 'resume' ? data.analysis.resume_markdown : data.analysis.portfolio_markdown}</pre><div className="report-method"><ShieldCheck size={16} /><p>{t.portfolioIntegrity}</p></div></section></div>}

          {view === 'settings' && <div className="view-stack">{data && overviewHeader}<div className="page-title"><div><span className="section-kicker">{t.workspacePreferences}</span><h1>{t.nav.settings}</h1><p>{t.settingsIntro}</p></div></div><section className="content-section language-settings"><div className="section-heading"><div><h2>{t.language}</h2><p>{t.languageDescription}</p></div><div className="language-switch large-switch"><button type="button" lang="en" aria-pressed={language === 'en'} onClick={() => setLanguage('en')}>{t.english}</button><button type="button" lang="fa" aria-pressed={language === 'fa'} onClick={() => setLanguage('fa')}>{t.persian}</button></div></div></section>{renderLLMSettings()}</div>}
        </main>
        <footer className="workspace-footer"><span>{t.footerEvidence}</span><span>GitIntel <bdi dir="ltr">·</bdi> {t.localAnalysis}</span></footer>
      </div>

      <nav className="mobile-nav" aria-label={t.primaryNavigation}>{navItems.filter((item) => ['analyze', 'dashboard', 'repositories', 'reports'].includes(item.id)).map(({ id, label, icon: Icon }) => <button key={id} className={view === id || (view === 'repository' && id === 'repositories') ? 'active' : ''} onClick={() => setView(id)}><Icon size={18} /><span>{label}</span></button>)}<button className={mobileMore || ['signals', 'timeline', 'portfolio', 'settings'].includes(view) ? 'active' : ''} onClick={() => setMobileMore((current) => !current)}><Menu size={18} /><span>{t.more}</span></button></nav>
      {mobileMore && <div className="mobile-more-menu">{navItems.filter((item) => ['signals', 'timeline', 'portfolio', 'settings'].includes(item.id)).map(({ id, label, icon: Icon }) => <button key={id} onClick={() => { setView(id); setMobileMore(false) }}><Icon size={17} />{label}</button>)}<button className="mobile-menu-close" onClick={() => setMobileMore(false)}><X size={16} />{t.close}</button></div>}
      {toast && <div className="toast" role="status"><Check size={16} />{toast}</div>}
    </div>
  )
}

export default App