import { ArrowUpRight, GitBranch, LoaderCircle, LockKeyhole, Search } from 'lucide-react'
import type { Messages } from './componentTypes'

interface AnalysisHeaderProps {
  username: string
  token: string
  loading: boolean
  compact?: boolean
  t: Messages
  onAnalyze: () => void
  onUsernameChange: (value: string) => void
  onTokenChange: (value: string) => void
}

export function AnalysisHeader({
  username,
  token,
  loading,
  compact = false,
  t,
  onAnalyze,
  onUsernameChange,
  onTokenChange,
}: AnalysisHeaderProps) {
  return (
    <form
      className={`analysis-form${compact ? ' compact-form' : ''}`}
      onSubmit={(event) => {
        event.preventDefault()
        onAnalyze()
      }}
    >
      <label className="field">
        <span>{t.username}</span>
        <div className="input-with-icon">
          <GitBranch size={compact ? 17 : 18} />
          <input
            dir="ltr"
            autoComplete={compact ? undefined : 'username'}
            value={username}
            onChange={(event) => onUsernameChange(event.target.value)}
            placeholder={compact ? undefined : t.usernamePlaceholder}
          />
          {!compact && <span className="input-prefix">github.com /</span>}
        </div>
      </label>
      {!compact && (
        <label className="field token-field">
          <span>{t.optionalToken}</span>
          <div className="input-with-icon">
            <LockKeyhole size={17} />
            <input
              dir="ltr"
              type="password"
              autoComplete="off"
              value={token}
              onChange={(event) => onTokenChange(event.target.value)}
              placeholder={t.tokenPlaceholder}
            />
          </div>
          <small>{t.tokenHelper}</small>
        </label>
      )}
      <button className={`button-primary${compact ? '' : ' analyze-button'}`} type="submit" disabled={loading || !username.trim()}>
        {loading ? <LoaderCircle className="spin" size={compact ? 16 : 17} /> : <Search size={compact ? 16 : 17} />}
        {loading ? t.analyzing : t.analyze}
        {!compact && <ArrowUpRight size={15} />}
      </button>
    </form>
  )
}
