import { useEffect, useMemo, useState } from 'react'
import './App.css'
import { getLLMStatus, interpretAnalysis, testLLMModel, type InterpretationResult, type LLMConfigStatus } from './llm'
import { messages, type Language } from './i18n'
import { analyzeProfile, copyText } from './services/api'
import type { AnalysisResult } from './types'

function App() {
  const [language, setLanguage] = useState<Language>(() => window.localStorage.getItem('gitintel-language') === 'fa' ? 'fa' : 'en')
  const t = messages[language]
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
  const [data, setData] = useState<AnalysisResult | null>(null)

  useEffect(() => {
    document.documentElement.lang = language
    document.documentElement.dir = language === 'fa' ? 'rtl' : 'ltr'
    window.localStorage.setItem('gitintel-language', language)
  }, [language])

  useEffect(() => {
    void getLLMStatus().then(setLLMConfig).catch(() => undefined)
  }, [])

  const summary = useMemo(() => {
    if (!data) return null
    return data.analysis.summary
  }, [data])

  const handleAnalyze = async () => {
    setLoading(true)
    setError('')
    try {
      const result = await analyzeProfile(username, token)
      setData(result)
      setInterpretation(null)
      setInterpretationError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to load GitHub profile.')
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
                : interpretationError
                  ? t.aiUnavailable
                  : ''

  const handleLLMTest = async () => {
    setLLMLoading(true)
    setLLMError('')
    try {
      const result = await testLLMModel(language)
      setLLMResult(result)
      setLLMConfig((current) => ({ ...current, credential_status: result.credential_status, provider_status: 'available' }))
    } catch (err) {
      const message = err instanceof Error ? err.message : ''
      if (message === 'INVALID_CREDENTIAL') {
        setLLMConfig((current) => ({ ...current, credential_status: 'invalid' }))
      } else if (message === 'RATE_LIMITED') {
        setLLMConfig((current) => ({ ...current, credential_status: 'temporarily unavailable' }))
      } else if (message === 'PAYMENT_REQUIRED') {
        setLLMConfig((current) => ({ ...current, provider_status: 'payment_required' }))
      }
      setLLMError(message === 'RATE_LIMITED' ? t.rateLimited : message === 'PAYMENT_REQUIRED' ? t.paymentRequired : message === 'INVALID_CREDENTIAL' ? t.invalidCredential : message || t.modelFailure)
    } finally {
      setLLMLoading(false)
    }
  }

  const credentialLabel = llmConfig.credential_status === 'active'
    ? t.active
    : llmConfig.credential_status === 'configured'
      ? t.configured
      : llmConfig.credential_status === 'invalid'
        ? t.invalid
        : llmConfig.credential_status === 'temporarily unavailable'
          ? t.temporarilyUnavailable
          : t.notConfigured
  const statusLabel = llmConfig.provider_status === 'payment_required'
    ? t.providerPaymentRequired
    : llmConfig.provider_status === 'rate_limited'
      ? t.rateLimited
      : !llmConfig.analysis_reports_enabled
        ? t.aiDisabled
        : llmResult
    ? llmResult.status === 'Success'
      ? t.success
      : llmResult.status === 'Rate Limited'
        ? t.rateLimited
        : llmResult.status === 'Testing'
          ? t.testingStatus
          : t.errorStatus
    : t.ready

  const copyResume = async () => {
    if (!data) return
    await copyText(data.analysis.resume_markdown)
  }

  const copyPortfolio = async () => {
    if (!data) return
    await copyText(data.analysis.portfolio_markdown)
  }

  const copyReport = async () => {
    if (!data) return
    await copyText(data.analysis.report_markdown)
  }

  const copyRepositoryLinks = async () => {
    if (!data) return
    await copyText(data.analysis.repository_links.join('\n'))
  }

  const downloadJSON = () => {
    if (!data) return
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${data.profile.username || 'gitintel'}-report.json`
    link.click()
    URL.revokeObjectURL(url)
  }

  const downloadMarkdown = () => {
    if (!data) return
    const blob = new Blob([data.analysis.report_markdown], { type: 'text/markdown' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${data.profile.username || 'gitintel'}-report.md`
    link.click()
    URL.revokeObjectURL(url)
  }

  const handlePrint = () => {
    window.print()
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="brand-block">
          <span className="brand-mark">gi</span>
          <div>
            <div className="brand-name">GitIntel</div>
            <small>{t.subtitle}</small>
          </div>
        </div>
        <div className="language-picker" role="group" aria-label={t.language}>
          <span>{t.language}</span>
          <div className="language-switch">
            <button type="button" lang="en" aria-pressed={language === 'en'} onClick={() => setLanguage('en')}>{t.english}</button>
            <button type="button" lang="fa" aria-pressed={language === 'fa'} onClick={() => setLanguage('fa')}>{t.persian}</button>
          </div>
        </div>
      </header>

      <main className="page">
        <section className="hero panel">
          <div className="hero-copy">
            <span className="eyebrow">{t.eyebrow}</span>
            <h1>{t.heroTitle}</h1>
            <p>{t.heroText}</p>

            <div className="input-row">
              <label className="field">
                <span>{t.username}</span>
                <input
                  dir="ltr"
                  value={username}
                  onChange={(e) => setUsername(e.target.value.trim())}
                  placeholder={t.usernamePlaceholder}
                />
              </label>
              <button className="primary" onClick={handleAnalyze} disabled={loading || !username}>
                {loading ? t.analyzing : t.analyze}
              </button>
            </div>

            <label className="field inline-field">
              <span>{t.optionalToken}</span>
              <input
                dir="ltr"
                type="password"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                placeholder={t.tokenPlaceholder}
              />
            </label>

            <div className="security-note">{t.security}</div>
          </div>
        </section>

        <section className="feature-grid">
          {t.features.map((feature) => (
            <div key={feature} className="feature-card panel">
              <span className="feature-name">{feature}</span>
            </div>
          ))}
        </section>

        {error && <div className="panel error-box">{error}</div>}

        <section className="panel block llm-box">
          <h3>{t.llmTitle}</h3>
          <div className="llm-grid">
            <div><span>{t.provider}</span><strong>{llmConfig.provider === 'huggingface' ? 'Hugging Face' : llmConfig.provider}</strong></div>
            <div><span>{t.model}</span><strong className="ltr-value">{llmConfig.model}</strong></div>
            <div><span>{t.language}</span><div className="language-switch inline-switch">
              <button type="button" lang="en" aria-pressed={language === 'en'} onClick={() => setLanguage('en')}>{t.english}</button>
              <button type="button" lang="fa" aria-pressed={language === 'fa'} onClick={() => setLanguage('fa')}>{t.persian}</button>
            </div></div>
            <div><span>{t.credential}</span><strong>{credentialLabel}</strong></div>
            <div><span>{t.status}</span><strong>{llmError === t.paymentRequired ? t.paymentRequired : llmError === t.rateLimited ? t.rateLimited : llmError === t.invalidCredential ? t.invalidCredential : llmError ? t.errorStatus : statusLabel}</strong></div>
            <div><span>{t.latency}</span><strong className="ltr-value">{llmResult ? `${llmResult.latency_ms}ms` : t.notTested}</strong></div>
            <div><span>{t.tokens}</span><strong className="ltr-value">{llmResult?.tokens ?? 'N/A'}</strong></div>
          </div>
          <div className="action-row">
            <button onClick={handleLLMTest} disabled={llmLoading || llmConfig.provider_status === 'payment_required'}>{llmLoading ? t.testing : t.testModel}</button>
          </div>
          {llmError && <p className="llm-error" role="alert">{llmError}</p>}
          {llmResult && (
            <div className="llm-output" dir="auto">
              <p dir="auto">{llmResult.response || llmResult.error || t.noResponse}</p>
            </div>
          )}
        </section>

        {data && (
          <div className="dashboard">
            <section className="panel profile-header">
              {data.profile.avatar_url && <img src={data.profile.avatar_url} alt={data.profile.username} />}
              <div>
                <h2>{data.profile.display_name || data.profile.username}</h2>
                <div className="meta-line"><bdi dir="ltr">@{data.profile.username}</bdi></div>
                <p>{data.profile.bio || t.profileFallback}</p>
              </div>
              <div className="profile-links">
                <a href={data.profile.profile_url} target="_blank" rel="noreferrer">{t.github}</a>
                {data.profile.blog && <a href={data.profile.blog} target="_blank" rel="noreferrer">{t.website}</a>}
              </div>
            </section>

            <section className="metrics-grid">
              {[{ label: t.metrics[0], value: data.profile.public_repos }, { label: t.metrics[1], value: data.profile.followers }, { label: t.metrics[2], value: data.profile.following }, { label: t.metrics[3], value: new Date(data.profile.created_at).getFullYear() }].map((metric) => (
                <div key={metric.label} className="panel metric-card">
                  <span>{metric.label}</span>
                  <strong className="ltr-value">{metric.value}</strong>
                </div>
              ))}
            </section>

            <section className="panel summary-box">
              <h3>{t.developerProfile}</h3>
              <p>{summary}</p>
            </section>

            <section className="panel block analysis-report">
              <h3>{t.analysisTitle}</h3>
              <p className="analysis-availability">{t.deterministicAvailable}</p>
              <pre className="deterministic-report" dir="ltr">{data.analysis.report_markdown}</pre>
              {interpretationErrorMessage && <p className="llm-error" role="status">{interpretationErrorMessage === t.aiUnavailable ? t.aiUnavailable : `${t.aiUnavailable} ${interpretationErrorMessage}`}</p>}
              {!llmConfig.analysis_reports_enabled && <p className="analysis-availability">{t.deterministicOnly} {t.interpretationDisabled}</p>}
              <div className="action-row">
                <button onClick={handleInterpret} disabled={interpretationLoading || !llmConfig.analysis_reports_enabled}>
                  {interpretationLoading ? t.generatingInterpretation : t.generateInterpretation}
                </button>
              </div>
              {interpretation && (
                <div className="interpretation-output" lang={interpretation.language} dir={interpretation.language === 'fa' ? 'rtl' : 'ltr'}>
                  <h4>{t.aiReport}</h4>
                  <p className="interpretation-meta" dir="ltr">{interpretation.provider} / {interpretation.model} / {interpretation.latency_ms}ms</p>
                  <p className="interpretation-text" dir="auto">{interpretation.report}</p>
                </div>
              )}
            </section>

            <section className="content-grid">
              <div className="panel block">
                <h3>{t.engineeringSignals}</h3>
                <div className="signal-list">
                  {data.analysis.signals.map((signal) => (
                    <div key={signal.name} className="signal-row">
                      <div className="signal-header">
                        <strong>{signal.name}</strong>
                        <span className="pill">{signal.level}</span>
                      </div>
                      <div className="signal-meta">{t.confidence}: <bdi dir="ltr">{signal.confidence}</bdi></div>
                      <div className="signal-meta">{t.metric}: <bdi dir="ltr">{Object.entries(signal.metric).map(([key, value]) => `${key}=${value}`).join(' · ')}</bdi></div>
                      <p>{signal.explanation}</p>
                      <ul>
                        {signal.evidence.map((item) => (
                          <li key={item}>{item}</li>
                        ))}
                      </ul>
                      <p className="signal-limitations">{t.limitations}: {signal.limitations.join(' ')}</p>
                    </div>
                  ))}
                </div>
              </div>

              <div className="panel block">
                <h3>{t.languageDistribution}</h3>
                <div className="language-list">
                  {data.analysis.language_distribution?.map((lang) => (
                    <div key={lang.name} className="language-row">
                      <div className="language-meta"><span>{lang.name}</span><span>{lang.percentage.toFixed(1)}%</span></div>
                      <div className="bar-track"><div className="bar-fill" style={{ width: `${lang.percentage}%` }} /></div>
                    </div>
                  )) || <span>{t.unavailable}</span>}
                </div>
              </div>
            </section>

            <section className="panel block">
              <h3>{t.featuredProjects}</h3>
              <div className="project-grid">
                {data.analysis.featured_projects.map((project) => (
                  <article key={project.repository.full_name} className="project-card">
                    <div className="project-head">
                      <strong className="ltr-value">{project.repository.name}</strong>
                      <a href={project.repository.url} target="_blank" rel="noreferrer">{t.github}</a>
                    </div>
                    <p>{project.repository.description || t.noDescription}</p>
                    <div className="project-meta">
                      <span>⭐ {project.repository.stars}</span>
                      <span>⑂ {project.repository.forks}</span>
                      <span>{project.repository.language || 'Multiple'}</span>
                    </div>
                    <ul>
                      {project.why.map((reason) => (
                        <li key={reason}>{reason}</li>
                      ))}
                    </ul>
                  </article>
                ))}
              </div>
            </section>

            <section className="panel block">
              <h3>{t.repoExplorer}</h3>
              <div className="repo-list">
                {data.repositories.map((repo) => (
                  <div key={repo.full_name} className="repo-item">
                    <div className="repo-title-row">
                      <strong className="ltr-value">{repo.name}</strong>
                      <a href={repo.url} target="_blank" rel="noreferrer">{t.github}</a>
                    </div>
                    <p>{repo.description || t.noDescription}</p>
                    <div className="project-meta">
                      <span className="ltr-value">{repo.language || t.languageUnknown}</span>
                      <span dir="ltr">⭐ {repo.stars}</span>
                      <span dir="ltr">⑂ {repo.forks}</span>
                      <span><bdi dir="ltr">{repo.open_issues}</bdi> {t.issues}</span>
                    </div>
                    <div className="topics ltr-value">{repo.topics.slice(0, 4).join(' • ') || t.noTopics}</div>
                  </div>
                ))}
              </div>
            </section>

            <section className="panel block export-box">
              <h3>{t.exportTools}</h3>
              <div className="action-row">
                <button onClick={copyResume}>{t.copyResume}</button>
                <button onClick={copyPortfolio}>{t.copyPortfolio}</button>
                <button onClick={copyReport}>{t.copyReport}</button>
                <button onClick={copyRepositoryLinks}>{t.exportLinks}</button>
                <button onClick={downloadJSON}>{t.downloadJSON}</button>
                <button onClick={downloadMarkdown}>{t.downloadMarkdown}</button>
                <button onClick={handlePrint}>{t.printPDF}</button>
              </div>
              <div className="export-list">
                {data.analysis.repository_links.map((link) => (
                  <div key={link} className="ltr-value">{link}</div>
                ))}
              </div>
            </section>
          </div>
        )}
      </main>
    </div>
  )
}

export default App
