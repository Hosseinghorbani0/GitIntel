import type { AnalysisResult, ApiResponse } from '../types'

const API_BASE = '/api'

export async function analyzeProfile(username: string, token = ''): Promise<AnalysisResult> {
  const response = await fetch(`${API_BASE}/analyze`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, token }),
  })

  const payload: ApiResponse<AnalysisResult> = await response.json()
  if (!response.ok || !payload.success || !payload.data) {
    throw new Error(payload.error?.message ?? 'Unable to analyze profile.')
  }

  return payload.data
}

export async function copyText(value: string) {
  if (!value) return
  await navigator.clipboard.writeText(value)
}
