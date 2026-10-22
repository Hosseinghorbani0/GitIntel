import { BookOpen, BriefcaseBusiness, Download, FileText, ShieldCheck, Sparkles } from 'lucide-react'
import type { AnalysisResult } from '../types'
import type { Messages } from './componentTypes'

export type PortfolioMode = 'resume' | 'portfolio' | 'profile_readme'

interface PortfolioViewProps {
  data: AnalysisResult
  mode: PortfolioMode
  onModeChange: (mode: PortfolioMode) => void
  onCopy: (value: string) => void
  onDownload: (filename: string, content: string, type: string) => void
  t: Messages
}

export function PortfolioView({ data, mode, onModeChange, onCopy, onDownload, t }: PortfolioViewProps) {
  let content = data.analysis.portfolio_markdown
  let title: string = t.portfolio
  if (mode === 'resume') {
    content = data.analysis.resume_markdown
    title = t.resume
  } else if (mode === 'profile_readme') {
    content = data.analysis.profile_readme_markdown || data.analysis.report_markdown
    title = t.profileReadme
  }

  const aiPrompt = data.analysis.ai_prompt_markdown || ''

  return (
    <>
      <div className="page-title"><div><span className="section-kicker">{t.evidenceBasedOutput}</span><h1>{t.nav.portfolio}</h1><p>{t.portfolioIntro}</p></div></div>
      <div className="segmented-control" role="tablist" aria-label={t.portfolioOutput}>
        <button role="tab" aria-selected={mode === 'resume'} onClick={() => onModeChange('resume')}><BriefcaseBusiness size={15} />{t.resume}</button>
        <button role="tab" aria-selected={mode === 'portfolio'} onClick={() => onModeChange('portfolio')}><BookOpen size={15} />{t.portfolio}</button>
        <button role="tab" aria-selected={mode === 'profile_readme'} onClick={() => onModeChange('profile_readme')}><Sparkles size={15} />{t.profileReadme}</button>
      </div>
      <section className="portfolio-document">
        <div className="portfolio-document-head">
          <div><span className="section-kicker">{t.observedOnGithub}</span><h2>{title}</h2></div>
          <div className="report-actions">
            {aiPrompt && (
              <button className="button-secondary" onClick={() => onCopy(aiPrompt)} title="Copy pre-engineered prompt for ChatGPT / Claude">
                <Sparkles size={15} />{t.copyAIPrompt}
              </button>
            )}
            <button className="button-secondary" onClick={() => onCopy(content)}><FileText size={15} />{t.copyText}</button>
            <button className="button-secondary" onClick={() => onDownload(`${data.profile.username}-${mode}.md`, content, 'text/markdown')}><Download size={15} />{t.downloadMarkdown}</button>
          </div>
        </div>
        <pre className="report-markdown" dir="ltr">{content}</pre>
        <div className="report-method"><ShieldCheck size={16} /><p>{t.portfolioIntegrity}</p></div>
      </section>
    </>
  )
}
