import type { Language } from '../i18n'
import type { Messages } from './componentTypes'

interface LanguageSwitchProps {
  language: Language
  t: Messages
  className?: string
  variant?: 'compact' | 'full'
  onChange: (language: Language) => void
}

export function LanguageSwitch({ language, t, className = '', variant = 'compact', onChange }: LanguageSwitchProps) {
  return (
    <div className={`language-switch ${className}`.trim()}>
      <button type="button" lang="en" aria-pressed={language === 'en'} onClick={() => onChange('en')}>{variant === 'full' ? t.english : 'EN'}</button>
      <button type="button" lang="fa" aria-pressed={language === 'fa'} onClick={() => onChange('fa')}>{variant === 'full' ? t.persian : 'فا'}</button>
    </div>
  )
}
