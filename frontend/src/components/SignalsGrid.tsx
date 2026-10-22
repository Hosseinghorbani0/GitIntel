import { CircleAlert } from 'lucide-react'
import type { Repository, Signal } from '../types'
import type { Messages } from './componentTypes'

interface SignalsGridProps {
  signals: Signal[]
  repositories: Repository[]
  language: 'en' | 'fa'
  t: Messages
  className?: string
}

export function SignalsGrid({ signals, repositories, language, t, className = '' }: SignalsGridProps) {
  return (
    <div className={`signal-grid ${className}`.trim()}>
      {signals.map((signal, index) => {
        const level = signal.level.toLowerCase()
        const stateClass = level.includes('insufficient') ? 'is-muted' : level.includes('limited') ? 'is-caution' : 'is-positive'
        const confidence = signal.confidence.toLowerCase()
        const confidenceLabel = confidence === 'high' ? t.confidenceHigh : confidence === 'medium' ? t.confidenceMedium : t.confidenceLow
        const levelLabel = level.includes('strong') ? t.signalLevelStrong
          : level.includes('moderate') ? t.signalLevelModerate
            : level.includes('insufficient') ? t.signalLevelInsufficient
              : level.includes('archived') ? t.signalLevelArchived : t.signalLevelLimited
        const signalTitle = signal.name in t.signalNameLabels
          ? t.signalNameLabels[signal.name as keyof typeof t.signalNameLabels]
          : signal.name
        const metricEntries = Object.entries(signal.metric)
        const metricMaximum = Math.max(1, ...metricEntries.map(([, value]) => value))
        const metricLabel = (key: string) => key in t.signalMetricLabels
          ? t.signalMetricLabels[key as keyof typeof t.signalMetricLabels]
          : key.replaceAll('_', ' ')
        const renderMetric = ([key, value]: [string, number]) => (
          <div className="signal-metric" key={key}>
            <div className="signal-metric-label">
              <span>{metricLabel(key)}</span>
              <strong><bdi dir="ltr">{new Intl.NumberFormat(language === 'fa' ? 'fa-IR' : 'en-US', { maximumFractionDigits: 1 }).format(value)}</bdi></strong>
            </div>
            <div className="metric-track" role="img" aria-label={`${metricLabel(key)}: ${value} ${t.relativeMetricScale}`}>
              <span style={{ width: `${value > 0 ? Math.max(5, value / metricMaximum * 100) : 0}%` }} />
            </div>
          </div>
        )

        return (
          <article key={signal.name} className={`signal-card ${stateClass}`}>
            <div className="signal-card-top"><span className="signal-index">0{index + 1}</span><span className="signal-state">{levelLabel}</span></div>
            <h3>{signalTitle}</h3>
            <p className="signal-explanation">{signal.explanation}</p>
            <div className="signal-measurements">
              <div className={`confidence-summary confidence-${confidence}`}>
                <span className="confidence-ring" role="img" aria-label={`${t.confidence}: ${confidenceLabel}`}><span>{confidenceLabel}</span></span>
                <span><small>{t.confidence}</small><strong>{confidenceLabel}</strong></span>
              </div>
              <div className="signal-metrics">
                {metricEntries.slice(0, 3).map(renderMetric)}
                {metricEntries.length > 3 && <details className="additional-metrics"><summary>{t.additionalMetrics}</summary>{metricEntries.slice(3).map(renderMetric)}</details>}
              </div>
              <p className="metric-context">{t.relativeMetricScale}</p>
            </div>
            <div className="signal-evidence-count"><span>{t.evidenceItems}</span><strong>{signal.evidence.length}</strong></div>
            {signal.evidence.length
              ? <ul className="evidence-list">{signal.evidence.slice(0, 3).map((item) => <li key={item}><bdi dir="auto">{item}</bdi></li>)}</ul>
              : <p className="quiet-empty">{t.insufficientEvidence}</p>}
            {signal.repositories && signal.repositories.length > 0 && (
              <div className="repo-tags signal-repositories">
                <span className="signal-repository-label">{t.supportingRepositories}</span>
                {signal.repositories.map((name) => {
                  const repository = repositories.find((item) => item.full_name === name)
                  return repository
                    ? <a key={name} href={repository.url} target="_blank" rel="noreferrer" dir="ltr">{repository.name}</a>
                    : <span key={name} dir="ltr">{name}</span>
                })}
              </div>
            )}
            {signal.limitations.length > 0 && (
              <details className="signal-limitations">
                <summary><CircleAlert size={13} />{t.viewLimitations} (<bdi dir="ltr">{signal.limitations.length}</bdi>)</summary>
                <p><bdi dir="auto">{signal.limitations.join(' ')}</bdi></p>
              </details>
            )}
          </article>
        )
      })}
    </div>
  )
}
