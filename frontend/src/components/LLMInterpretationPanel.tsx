import { Activity, LoaderCircle } from 'lucide-react'
import type { InterpretationResult } from '../llm'
import type { Messages } from './componentTypes'

interface LLMInterpretationPanelProps {
  interpretation: InterpretationResult | null
  loading: boolean
  enabled: boolean
  error: string
  statusLabel: string
  onGenerate: () => void
  t: Messages
}

export function LLMInterpretationPanel({
  interpretation,
  loading,
  enabled,
  error,
  statusLabel,
  onGenerate,
  t,
}: LLMInterpretationPanelProps) {
  return (
    <section className="ai-availability">
      <div className="ai-label"><span className="ai-indicator" /><div><strong>{t.aiInterpretation}</strong><p>{t.aiSeparateLayer}</p></div></div>
      <span className="status-tag status-unavailable"><span />{statusLabel}</span>
      <p className="ai-explanation">{error || (!enabled ? t.interpretationDisabled : t.aiAvailable)}</p>
      <button className="button-secondary" onClick={onGenerate} disabled={loading || !enabled}>
        {loading ? <LoaderCircle className="spin" size={15} /> : <Activity size={15} />}
        {loading ? t.generatingInterpretation : t.generateInterpretation}
      </button>
      {interpretation && (
        <div className="interpretation-output" lang={interpretation.language} dir={interpretation.language === 'fa' ? 'rtl' : 'ltr'}>
          <div className="interpretation-meta" dir="ltr">
            {interpretation.status === 'streaming' ? t.generatingInterpretation : `${interpretation.provider} / ${interpretation.model} / ${interpretation.latency_ms}ms`}
          </div>
          <p>{interpretation.report}</p>
        </div>
      )}
    </section>
  )
}
