import { ArrowUpRight, GitBranch, Search } from 'lucide-react'
import type { AnalysisResult } from '../types'
import type { Messages } from './componentTypes'

interface ProfileHeroProps {
  data: AnalysisResult
  analyzedAt: string
  language: 'en' | 'fa'
  t: Messages
  onNewAnalysis: () => void
}

export function ProfileHero({ data, analyzedAt, language, t, onNewAnalysis }: ProfileHeroProps) {
  return (
    <section className="identity-strip">
      {data.profile.avatar_url
        ? <img className="avatar" src={data.profile.avatar_url} alt="" />
        : <div className="avatar avatar-fallback"><GitBranch size={23} /></div>}
      <div className="identity-copy">
        <div className="identity-name-row">
          <h1><bdi dir="auto">{data.profile.display_name || data.profile.username}</bdi></h1>
          <span className="status-tag"><span />{t.analysisReady}</span>
        </div>
        <div className="identity-meta">
          <bdi dir="ltr">@{data.profile.username}</bdi><span>·</span>
          <span>{t.analyzedAt} <bdi dir="ltr">{analyzedAt
            ? new Date(analyzedAt).toLocaleString(language === 'fa' ? 'fa-IR' : 'en-US', { dateStyle: 'medium', timeStyle: 'short' })
            : t.unavailable}</bdi></span>
        </div>
        {data.profile.bio && <p>{data.profile.bio}</p>}
      </div>
      <div className="identity-actions">
        <a className="quiet-button" href={data.profile.profile_url} target="_blank" rel="noreferrer">
          <GitBranch size={15} />{t.viewGithub}<ArrowUpRight size={13} />
        </a>
        <button className="quiet-button" onClick={onNewAnalysis}>
          <Search size={15} />{t.newAnalysis}
        </button>
      </div>
    </section>
  )
}
