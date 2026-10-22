import { useEffect, useRef, useState, type FormEvent } from 'react'
import {
  Activity, ArrowUpRight, BookOpen, BriefcaseBusiness, Check,
  CircleAlert, Clock3, Code2, Download, FileText, GitFork, Globe2, LayoutDashboard,
  LoaderCircle, Menu, Network, Search, Settings2, ShieldCheck,
  Star, X,
} from 'lucide-react'
import './App.css'
import { getLLMStatus, streamInterpretAnalysis, testLLMModel, type InterpretationResult, type LLMConfigStatus } from './llm'
import { messages, type Language } from './i18n'
import { analyzeProfile, copyText } from './services/api'
import type { AnalysisResult, Repository } from './types'
import { AnalysisHeader } from './components/AnalysisHeader'
import { LanguageSwitch } from './components/LanguageSwitch'
import { LLMInterpretationPanel } from './components/LLMInterpretationPanel'
import { PortfolioView, type PortfolioMode } from './components/PortfolioView'
import { ProfileHero } from './components/ProfileHero'
import { RepositoryExplorer } from './components/RepositoryExplorer'
import { SignalsGrid } from './components/SignalsGrid'

type ViewKey = 'analyze' | 'dashboard' | 'repositories' | 'signals' | 'timeline' | 'reports' | 'portfolio' | 'settings'

const viewKeys: ViewKey[] = ['analyze', 'dashboard', 'repositories', 'signals', 'timeline', 'reports', 'portfolio', 'settings']

function viewFromHash(hash: string): ViewKey {
  const route = hash.replace(/^#\/?/, '')
  return viewKeys.includes(route as ViewKey) ? route as ViewKey : 'analyze'
}

function AppWorkspace() {
  const [language, setLanguage] = useState<Language>(() => window.localStorage.getItem('gitintel-language') === 'fa' ? 'fa' : 'en')
  const t = messages[language]
  const [initialDeepLinkUsername] = useState(() => new URLSearchParams(window.location.search).get('u')?.trim() || '')
  const [initialAnalysisFailedMessage] = useState(() => t.analysisFailed)
  const [view, setView] = useState<ViewKey>(() => viewFromHash(window.location.hash))
  const [mobileMore, setMobileMore] = useState(false)
  const [username, setUsername] = useState(() => new URLSearchParams(window.location.search).get('u')?.trim() || 'Hosseinghorbani0')
  const [token, setToken] = useState('')
  const [loading, setLoading] = useState(() => Boolean(initialDeepLinkUsername))
  const analysisController = useRef<AbortController | null>(null)
  const [llmLoading, setLLMLoading] = useState(false)
  const [interpretationLoading, setInterpretationLoading] = useState(false)
  const [interpretation, setInterpretation] = useState<InterpretationResult | null>(null)
  const [interpretationError, setInterpretationError] = useState('')
  const interpretationController = useRef<AbortController | null>(null)
  const [llmResult, setLLMResult] = useState<{ provider: string; model: string; status: string; latency_ms: number; tokens?: number; response: string; error?: string } | null>(null)
  const [llmConfig, setLLMConfig] = useState<LLMConfigStatus>({ provider: 'huggingface', model: 'Qwen/Qwen3.8-2.4T-A95B:novita', credential_status: 'not configured', provider_status: 'unverified', analysis_reports_enabled: false })
  const [llmError, setLLMError] = useState('')
  const [error, setError] = useState('')
  const [toast, setToast] = useState('')
  const [data, setData] = useState<AnalysisResult | null>(null)
  const [analyzedAt, setAnalyzedAt] = useState('')
  const [selectedRepository, setSelectedRepository] = useState<Repository | null>(null)
  const repositoryDialogRef = useRef<HTMLDialogElement>(null)
  const [portfolioMode, setPortfolioMode] = useState<PortfolioMode>('profile_readme')

  useEffect(() => {
    document.documentElement.lang = language
    document.documentElement.dir = language === 'fa' ? 'rtl' : 'ltr'
    window.localStorage.setItem('gitintel-language', language)
  }, [language])

  useEffect(() => {
    if (!initialDeepLinkUsername) return
    const controller = new AbortController()
    analysisController.current?.abort()
    analysisController.current = controller
    void analyzeProfile(initialDeepLinkUsername, '', controller.signal)
      .then((result) => {
        if (analysisController.current !== controller) return
        setData(result)
        setAnalyzedAt(new Date().toISOString())
        setView('dashboard')
        const url = new URL(window.location.href)
        url.searchParams.set('u', result.profile.username)
        window.history.replaceState(window.history.state, '', url)
      })
      .catch((err: unknown) => {
        if (!controller.signal.aborted && analysisController.current === controller) {
          setError(err instanceof Error ? err.message : initialAnalysisFailedMessage)
        }
      })
      .finally(() => {
        if (analysisController.current === controller) {
          analysisController.current = null
          setLoading(false)
        }
      })
    return () => {
      controller.abort()
      if (analysisController.current === controller) analysisController.current = null
    }
  }, [initialDeepLinkUsername, initialAnalysisFailedMessage])

  useEffect(() => () => {
    analysisController.current?.abort()
    interpretationController.current?.abort()
  }, [])

  useEffect(() => {
    const hash = window.location.hash
    const route = hash.replace(/^#\/?/, '')
    const target = `#/${view}`
    if (hash === target) return
    if (!hash || !viewKeys.includes(route as ViewKey)) {
      window.history.replaceState({ view }, '', target)
    } else if (route !== view) {
      window.history.pushState({ view }, '', target)
    }
  }, [view])

  useEffect(() => {
    const syncView = () => {
      setView(viewFromHash(window.location.hash))
      setSelectedRepository(null)
    }
    window.addEventListener('popstate', syncView)
    window.addEventListener('hashchange', syncView)
    return () => {
      window.removeEventListener('popstate', syncView)
      window.removeEventListener('hashchange', syncView)
    }
  }, [])

  useEffect(() => {
    const dialog = repositoryDialogRef.current
    if (!dialog) return
    if (selectedRepository && !dialog.open) dialog.showModal()
    if (!selectedRepository && dialog.open) dialog.close()
  }, [selectedRepository])

  useEffect(() => {
    void getLLMStatus().then(setLLMConfig).catch(() => undefined)
  }, [])

  useEffect(() => {
    if (!toast) return
    const timeout = window.setTimeout(() => setToast(''), 2600)
    return () => window.clearTimeout(timeout)
  }, [toast])

  const repositoryCollaborationSummary = (repo: Repository) => {
    const facts: string[] = []
    if (repo.pull_request_count_status === 'verified') {
      facts.push(`${t.openIssuesLabel}: ${repo.open_issues_count ?? 0}`, `${t.openPullRequestsLabel}: ${repo.open_pull_requests ?? 0}`)
    }
    if (repo.contributors_status === 'verified') {
      facts.push(`${t.contributorCountLabel}: ${repo.contributors_count ?? 0}`)
    }
    return facts.length ? facts.join(' · ') : t.collaborationCountsUnverified
  }

  const repositoryDocumentationSummary = (repo: Repository) => {
    const facts: string[] = []
    if (repo.readme_status === 'verified_present' || repo.has_readme) facts.push(t.readmeEvidence)
    else if (repo.readme_status === 'verified_absent') facts.push(t.readmeNotObserved)
    else facts.push(t.readmeNotVerified)
    if (repo.has_pages) facts.push(t.pagesEnabled)
    if (repo.has_wiki) facts.push(t.wikiEnabled)
    if (repo.has_license) facts.push(t.licensePresent)
    return facts.join(' · ')
  }

  const repositoryReleaseSummary = (repo: Repository) => {
    if (repo.release_status === 'verified_present' || repo.has_releases) return `${t.releaseMetadataObserved} · ${t.releaseCountLabel}: ${repo.release_count ?? 0}`
    if (repo.release_status === 'verified_absent') return t.releaseMetadataNotObserved
    return t.releaseCountNotVerified
  }

  const handleAnalyze = async (event?: FormEvent) => {
    event?.preventDefault()
    const requestedUsername = username.trim()
    if (!requestedUsername) return
    analysisController.current?.abort()
    interpretationController.current?.abort()
    interpretationController.current = null
    setInterpretationLoading(false)
    setInterpretation(null)
    setInterpretationError('')
    const controller = new AbortController()
    analysisController.current = controller
    setLoading(true)
    setError('')
    try {
      const result = await analyzeProfile(requestedUsername, token, controller.signal)
      if (analysisController.current !== controller) return
      setData(result)
      setAnalyzedAt(new Date().toISOString())
      setInterpretation(null)
      setInterpretationError('')
      setSelectedRepository(null)
      setView('dashboard')
      const url = new URL(window.location.href)
      url.searchParams.set('u', result.profile.username)
      window.history.replaceState(window.history.state, '', url)
      setToast(t.analysisComplete)
    } catch (err) {
      if (!controller.signal.aborted && analysisController.current === controller) {
        setError(err instanceof Error ? err.message : t.analysisFailed)
      }
    } finally {
      if (analysisController.current === controller) {
        analysisController.current = null
        setLoading(false)
      }
    }
  }

  const handleLanguageChange = (nextLanguage: Language) => {
    if (nextLanguage === language) return
    interpretationController.current?.abort()
    interpretationController.current = null
    setInterpretationLoading(false)
    setInterpretation(null)
    setInterpretationError('')
    setLanguage(nextLanguage)
  }

  const handleInterpret = async () => {
    if (!data) return
    interpretationController.current?.abort()
    const controller = new AbortController()
    interpretationController.current = controller
    setInterpretationLoading(true)
    setInterpretationError('')
    setInterpretation(null)
    try {
      const result = await streamInterpretAnalysis(data.profile.username, language, (report) => {
        setInterpretation({
          status: 'streaming',
          language,
          provider: llmConfig.provider,
          model: llmConfig.model,
          report,
          latency_ms: 0,
          tokens: 0,
        })
      }, controller.signal)
      setInterpretation(result)
    } catch (err) {
      if (!controller.signal.aborted) {
        setInterpretation(null)
        setInterpretationError(err instanceof Error ? err.message : 'LLM_PROVIDER_ERROR')
      }
    } finally {
      if (interpretationController.current === controller) {
        interpretationController.current = null
        setInterpretationLoading(false)
      }
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
  const activeLabel = navItems.find((item) => item.id === view)?.label ?? t.nav.dashboard
  const overviewHeader = data && (
    <ProfileHero
      data={data}
      analyzedAt={analyzedAt}
      language={language}
      t={t}
      onNewAnalysis={() => { setView('analyze'); window.scrollTo({ top: 0, behavior: 'smooth' }) }}
    />
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
          {navItems.map(({ id, label, icon: Icon }) => <button key={id} className={`nav-link ${view === id ? 'active' : ''}`} aria-current={view === id ? 'page' : undefined} onClick={() => { setView(id); setMobileMore(false) }}><Icon size={17} strokeWidth={1.8} /><span>{label}</span>{id === 'dashboard' && data && <span className="nav-count">{data.analysis.signals.length}</span>}</button>)}
        </nav>
        <div className="sidebar-spacer" />
        <div className="connection-state"><span className="connection-dot" /><span>{t.githubPublicData}</span></div>
        <div className="sidebar-language"><Globe2 size={15} /><span>{t.language}</span><LanguageSwitch language={language} onChange={handleLanguageChange} t={t} /></div>
        <div className="sidebar-user"><div className="user-avatar">GI</div><div><strong>{data?.profile.username || 'GitIntel'}</strong><span>{data ? t.localAnalysis : t.personalWorkspace}</span></div></div>
      </aside>

      <div className="workspace-main">
        <header className="mobile-topbar"><button className="mobile-brand" onClick={() => setView('analyze')}><span className="brand-symbol"><span /><span /><span /></span><b>GitIntel<span className="brand-period">.</span></b></button><LanguageSwitch language={language} onChange={handleLanguageChange} t={t} /></header>
        <header className="workspace-topbar"><div className="breadcrumb"><span>{t.workspace}</span><span>/</span><strong>{activeLabel}</strong></div>        <div className="topbar-context">{data ? <><span className="context-dot" /><bdi dir="ltr">{data.profile.username}</bdi><span className="context-divider" /><span><bdi dir="ltr">{data.repositories.length}</bdi> {t.repositoriesAnalyzed}</span></> : <span>{t.evidenceWorkspace}</span>}</div></header>
        <main className="main-content">
          {error && <div className="global-error" role="alert"><div className="error-icon"><CircleAlert size={18} /></div><div><strong>{t.analysisFailed}</strong><p>{error}</p><button onClick={() => setError('')}>{t.dismiss}</button></div></div>}
          {view === 'analyze' && <>
            {!data ? <section className="analyze-view">
              <div className="analyze-intro"><span className="section-kicker"><span className="kicker-mark" />{t.evidenceWorkspace}</span><h1>{t.heroTitle}</h1><p>{t.heroText}</p></div>
              <AnalysisHeader
                username={username}
                token={token}
                loading={loading}
                t={t}
                onAnalyze={() => void handleAnalyze()}
                onUsernameChange={setUsername}
                onTokenChange={setToken}
              />
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
              <AnalysisHeader
                username={username}
                token={token}
                loading={loading}
                compact
                t={t}
                onAnalyze={() => void handleAnalyze()}
                onUsernameChange={setUsername}
                onTokenChange={setToken}
              />
              {loading && <div className="loading-panel" role="status"><div className="loading-heading"><LoaderCircle className="spin" size={19} /><div><strong>{t.analysisInProgress}</strong><span>{t.analysisProgressDetail}</span></div></div><div className="loading-track"><span /></div></div>}
            </>}
          </>}

          {view !== 'analyze' && !data && emptyState(t.noAnalysisTitle, t.noAnalysisDescription)}

          {data && view === 'dashboard' && <div className="view-stack">
            {overviewHeader}
            <div className="overview-heading"><div><span className="section-kicker">{t.observedOverview}</span><h1>{t.developerProfile}</h1></div><p>{t.deterministicAvailable}</p></div>
            <section className="overview-grid">
              <article className="overview-summary"><div className="summary-label"><span className="summary-mark"><Activity size={16} /></span>{t.evidenceSummary}<span className="deterministic-badge">{t.deterministic}</span></div><p><bdi dir="auto">{data.analysis.summary || t.insufficientEvidence}</bdi></p><div className="summary-foot"><span><ShieldCheck size={14} />{t.evidenceOnly}</span><button onClick={() => setView('reports')}>{t.openReport}<ArrowUpRight size={14} /></button></div></article>
              <div className="stat-stack">
                <div className="stat-cell"><span>{t.repositoriesAnalyzed}</span><strong><bdi dir="ltr">{data.repositories.length}</bdi></strong><small>{t.publicRepositoriesCount} <bdi dir="ltr">{data.profile.public_repos}</bdi></small></div>
                <div className="stat-cell"><span>{t.languagesObserved}</span><strong><bdi dir="ltr">{data.analysis.language_distribution?.length ?? 0}</bdi></strong><small>{data.analysis.language_distribution_basis === 'code_bytes_sampled' ? <>{t.codeBytesSample} · <bdi dir="ltr">{data.analysis.language_repositories_sampled}</bdi> {t.repositoriesSampled}</> : t.repositoryFrequencyFallback}</small></div>
                <div className="stat-cell"><span>{t.profileSince}</span><strong><bdi dir="ltr">{new Date(data.profile.created_at).getFullYear()}</bdi></strong><small>{t.githubAccount}</small></div>
              </div>
            </section>
            <section className="content-section"><div className="section-heading"><div><span className="section-kicker">{t.observedSignals}</span><h2>{t.engineeringSignals}</h2><p>{t.signalsIntro}</p></div><button className="text-link" onClick={() => setView('signals')}>{t.allSignals}<ArrowUpRight size={14} /></button></div><SignalsGrid signals={data.analysis.signals.slice(0, 3)} repositories={data.repositories} language={language} t={t} /><div className="signals-footer"><span>{t.noDeveloperScore}</span><button onClick={() => setView('signals')}>{t.viewAllSignals}<ArrowUpRight size={14} /></button></div></section>
            <section className="content-section compact-section"><div className="section-heading"><div><span className="section-kicker">{t.projects}</span><h2>{t.featuredProjects}</h2></div><button className="text-link" onClick={() => setView('repositories')}>{t.browseRepositories}<ArrowUpRight size={14} /></button></div>{data.analysis.featured_projects.length ? <div className="repo-grid featured-grid">{data.analysis.featured_projects.slice(0, 3).map((project) => <article className="featured-row" key={project.repository.full_name}><div className="featured-main"><span className="repo-icon"><Code2 size={16} /></span><div><strong dir="ltr">{project.repository.name}</strong><p><bdi dir="auto">{project.repository.description || t.noDescription}</bdi></p></div></div><div className="repo-meta"><span>{project.repository.language || t.languageUnknown}</span><span><Star size={13} /><bdi dir="ltr">{project.repository.stars}</bdi></span></div>{project.why.length > 0 && <small><bdi dir="auto">{project.why[0]}</bdi></small>}</article>)}</div> : <p className="quiet-empty">{t.noFeaturedProjects}</p>}</section>
          </div>}

          {data && <div className="view-stack" hidden={view !== 'repositories'}>
            {overviewHeader}
            <RepositoryExplorer
              data={data}
              analyzedAt={analyzedAt}
              language={language}
              t={t}
              onSelectRepository={setSelectedRepository}
            />
          </div>}

          {data && view === 'signals' && <div className="view-stack">{overviewHeader}<div className="page-title"><div><span className="section-kicker">{t.observableEvidence}</span><h1>{t.engineeringSignals}</h1><p>{t.signalsIntro}</p></div><span className="deterministic-badge">{t.noDeveloperScore}</span></div><SignalsGrid signals={data.analysis.signals} repositories={data.repositories} language={language} t={t} className="signal-grid-all" /><section className="content-section attention-section"><div className="section-heading"><div><span className="section-kicker">{t.inspectWithContext}</span><h2>{t.attentionSignals}</h2><p>{t.attentionIntro}</p></div></div>{data.analysis.signals.filter((signal) => /limited|insufficient/i.test(signal.level)).length ? <div className="attention-list">{data.analysis.signals.filter((signal) => /limited|insufficient/i.test(signal.level)).map((signal) => <article className="attention-row" key={signal.name}><span className="attention-icon"><CircleAlert size={16} /></span><div><strong>{signal.name}</strong><p>{signal.explanation}</p><small>{signal.limitations.join(' ') || t.cannotConclude}</small></div><span className="signal-state">{signal.level}</span></article>)}</div> : <div className="empty-inline"><Check size={18} /><strong>{t.noAttentionSignals}</strong><span>{t.noAttentionDescription}</span></div>}</section></div>}

          {data && view === 'timeline' && <div className="view-stack">{overviewHeader}<div className="page-title"><div><span className="section-kicker">{t.observedHistory}</span><h1>{t.timeline}</h1><p>{t.timelineIntro}</p></div></div>{data.repositories.length ? <><div className="timeline-list">{data.repositories.flatMap((repo) => [{ key: `${repo.full_name}-created`, date: repo.created_at, title: t.repositoryCreated, repo }, { key: `${repo.full_name}-updated`, date: repo.updated_at, title: t.repositoryUpdated, repo }]).sort((left, right) => Date.parse(right.date) - Date.parse(left.date)).slice(0, 40).map((event) => <article className="timeline-event" key={event.key}><span className="timeline-marker" /><time dateTime={event.date}>{new Date(event.date).toLocaleDateString(language === 'fa' ? 'fa-IR' : 'en-US', { dateStyle: 'medium' })}</time><div><strong>{event.title}</strong><button dir="ltr" aria-haspopup="dialog" onClick={() => setSelectedRepository(event.repo)}>{event.repo.full_name}<ArrowUpRight size={13} /></button></div></article>)}</div><p className="timeline-note"><CircleAlert size={15} />{t.timelineLimitation}</p></> : <div className="empty-inline"><Clock3 size={18} /><strong>{t.noTimelineData}</strong><span>{t.noTimelineDescription}</span></div>}</div>}

          {data && view === 'reports' && <div className="view-stack report-view">
            {overviewHeader}
            <div className="page-title"><div><span className="section-kicker">{t.deterministic}</span><h1>{t.analysisTitle}</h1><p>{t.reportIntro}</p></div><div className="report-actions"><button className="button-secondary" onClick={() => window.print()}><FileText size={15} />{t.printPDF}</button><button className="button-secondary" onClick={() => downloadFile(`${data.profile.username}-report.md`, data.analysis.report_markdown, 'text/markdown')}><Download size={15} />{t.downloadMarkdown}</button></div></div>
            <article className="report-document"><div className="report-document-head"><div><span className="section-kicker">GITINTEL / ENGINEERING EVIDENCE</span><h2>{data.profile.display_name || data.profile.username}</h2><p><bdi dir="ltr">@{data.profile.username}</bdi> · {t.analyzedAt} <bdi dir="ltr">{analyzedAt ? new Date(analyzedAt).toLocaleDateString(language === 'fa' ? 'fa-IR' : 'en-US') : t.unavailable}</bdi></p></div><span className="deterministic-badge">{t.deterministic}</span></div><pre className="report-markdown" dir="ltr">{data.analysis.report_markdown}</pre><div className="report-method"><ShieldCheck size={16} /><p>{t.reportMethodology}</p></div></article>
            <LLMInterpretationPanel
              interpretation={interpretation}
              loading={interpretationLoading}
              enabled={llmConfig.analysis_reports_enabled}
              error={interpretationErrorMessage || (!llmConfig.analysis_reports_enabled
                ? llmConfig.provider_status === 'payment_required' ? t.paymentRequired : t.interpretationDisabled
                : t.aiAvailable)}
              statusLabel={statusLabel}
              onGenerate={() => void handleInterpret()}
              t={t}
            />
            <div className="report-actions bottom-actions"><button className="button-secondary" onClick={() => void copyValue(data.analysis.report_markdown)}><FileText size={15} />{t.copyReport}</button><button className="button-secondary" onClick={() => downloadFile(`${data.profile.username}-report.json`, JSON.stringify(data, null, 2), 'application/json')}><Download size={15} />{t.downloadJSON}</button></div>
          </div>}

          {data && view === 'portfolio' && <div className="view-stack">{overviewHeader}<PortfolioView data={data} mode={portfolioMode} onModeChange={setPortfolioMode} onCopy={(value) => void copyValue(value)} onDownload={downloadFile} t={t} /></div>}

          {view === 'settings' && <div className="view-stack">{data && overviewHeader}<div className="page-title"><div><span className="section-kicker">{t.workspacePreferences}</span><h1>{t.nav.settings}</h1><p>{t.settingsIntro}</p></div></div><section className="content-section language-settings"><div className="section-heading"><div><h2>{t.language}</h2><p>{t.languageDescription}</p></div><LanguageSwitch language={language} onChange={handleLanguageChange} t={t} className="large-switch" variant="full" /></div></section>{renderLLMSettings()}</div>}
        </main>
        <footer className="workspace-footer"><span>{t.footerEvidence}</span><span>GitIntel <bdi dir="ltr">·</bdi> {t.localAnalysis}</span></footer>
      </div>

      {data && <dialog
        ref={repositoryDialogRef}
        className="repository-drawer"
        aria-labelledby="repository-drawer-title"
        onClose={() => setSelectedRepository(null)}
        onClick={(event) => { if (event.target === event.currentTarget) event.currentTarget.close() }}
      >
        {selectedRepository && <>
          <header className="drawer-header">
            <div><span className="section-kicker">{t.repositoryInspection}</span><h2 id="repository-drawer-title" dir="ltr">{selectedRepository.full_name}</h2></div>
            <button className="drawer-close" type="button" aria-label={t.closeRepositoryDetails} onClick={() => setSelectedRepository(null)}><X size={18} /></button>
          </header>
          <div className="drawer-content">
            <p className="drawer-description">{selectedRepository.description || t.noDescription}</p>
            <div className="drawer-actions"><a className="quiet-button" href={selectedRepository.url} target="_blank" rel="noreferrer">{t.viewGithub}<ArrowUpRight size={14} /></a>{/^https?:\/\/\S+$/i.test(selectedRepository.homepage) && <a className="quiet-button" href={selectedRepository.homepage} target="_blank" rel="noreferrer">{t.homepage}<ArrowUpRight size={14} /></a>}</div>
            <div className="drawer-meta">
              <span>{selectedRepository.language || t.languageUnknown}</span>
              <span><Star size={13} />{selectedRepository.stars}</span>
              <span><GitFork size={13} />{selectedRepository.forks}</span>
              <span>{selectedRepository.license || t.licenseUnknown}</span>
              {selectedRepository.archived && <span>{t.archived}</span>}
              {selectedRepository.fork && <span>{t.fork}</span>}
            </div>
            <section className="drawer-counts" aria-label={t.collaborationSection}>
              <div><span>{t.openIssuesLabel}</span><strong>{selectedRepository.pull_request_count_status === 'verified' ? selectedRepository.open_issues_count ?? 0 : t.notVerifiedValue}</strong></div>
              <div><span>{t.openPullRequestsLabel}</span><strong>{selectedRepository.pull_request_count_status === 'verified' ? selectedRepository.open_pull_requests ?? 0 : t.notVerifiedValue}</strong></div>
              <div><span>{t.contributorCountLabel}</span><strong>{selectedRepository.contributors_status === 'verified' ? selectedRepository.contributors_count ?? 0 : t.notVerifiedValue}</strong></div>
            </section>
            <div className="drawer-detail-grid">
              {[
                { title: t.activitySection, text: `${t.created} ${new Date(selectedRepository.created_at).toLocaleDateString(language === 'fa' ? 'fa-IR' : 'en-US')} · ${t.updated} ${new Date(selectedRepository.updated_at).toLocaleDateString(language === 'fa' ? 'fa-IR' : 'en-US')}`, icon: Clock3 },
                { title: t.testingSection, text: selectedRepository.testing_artifacts_status === 'verified_present' ? `${t.testingArtifactsObserved}: ${selectedRepository.testing_artifacts?.join(', ')}` : selectedRepository.testing_artifacts_status === 'verified_absent' ? t.testingArtifactsNotObserved : selectedRepository.testing_artifacts_check_attempted ? t.testingArtifactsUnverified : t.testingNotCollected, icon: Check },
                { title: t.documentationSection, text: repositoryDocumentationSummary(selectedRepository), icon: BookOpen },
                { title: t.collaborationSection, text: repositoryCollaborationSummary(selectedRepository), icon: Network },
                { title: t.deliverySection, text: repositoryReleaseSummary(selectedRepository), icon: ArrowUpRight },
                { title: t.limitations, text: t.repositoryLimitations, icon: ShieldCheck },
              ].map(({ title, text, icon: Icon }) => <article className="detail-panel" key={title}><div className="detail-panel-title"><Icon size={16} /><h3>{title}</h3></div><p dir="auto">{text}</p></article>)}
            </div>
            {(selectedRepository.topics?.length ?? 0) > 0 && <div className="repo-tags drawer-topics">{selectedRepository.topics.map((topic) => <span key={topic}>{topic}</span>)}</div>}
          </div>
        </>}
      </dialog>}

      <nav className="mobile-nav" aria-label={t.primaryNavigation}>{navItems.filter((item) => ['analyze', 'dashboard', 'repositories', 'reports'].includes(item.id)).map(({ id, label, icon: Icon }) => <button key={id} className={view === id ? 'active' : ''} aria-current={view === id ? 'page' : undefined} onClick={() => setView(id)}><Icon size={18} /><span>{label}</span></button>)}<button className={mobileMore || ['signals', 'timeline', 'portfolio', 'settings'].includes(view) ? 'active' : ''} aria-expanded={mobileMore} aria-controls="mobile-more-menu" onClick={() => setMobileMore((current) => !current)}><Menu size={18} /><span>{t.more}</span></button></nav>
      {mobileMore && <div id="mobile-more-menu" className="mobile-more-menu">{navItems.filter((item) => ['signals', 'timeline', 'portfolio', 'settings'].includes(item.id)).map(({ id, label, icon: Icon }) => <button key={id} aria-current={view === id ? 'page' : undefined} onClick={() => { setView(id); setMobileMore(false) }}><Icon size={17} />{label}</button>)}<button className="mobile-menu-close" onClick={() => setMobileMore(false)}><X size={16} />{t.close}</button></div>}
      {toast && <div className="toast" role="status"><Check size={16} />{toast}</div>}
    </div>
  )
}

export default AppWorkspace