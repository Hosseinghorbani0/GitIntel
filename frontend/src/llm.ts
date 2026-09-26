export interface LLMStatusResponse {
  provider: string
  model: string
  status: string
  credential_status: string
  latency_ms: number
  tokens?: number
  response: string
  error?: string
  created_at?: string
}

export interface LLMConfigStatus {
  provider: string
  model: string
  credential_status: string
  provider_status: 'available' | 'payment_required' | 'rate_limited' | 'unavailable' | 'unverified' | 'disabled'
  analysis_reports_enabled: boolean
}

export interface InterpretationResult {
  status: string
  language: 'en' | 'fa'
  provider: string
  model: string
  report: string
  latency_ms: number
  tokens: number
}

export interface LLMTestResponse {
  success: boolean
  data?: LLMStatusResponse
  error?: {
    code: string
    message: string
  }
}

export async function testLLMModel(language: 'en' | 'fa'): Promise<LLMStatusResponse> {
  const response = await fetch('/api/llm/test', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ language }),
  })

  const payload: LLMTestResponse = await response.json()
  if (!response.ok || !payload.success || !payload.data) {
    const code = payload.error?.code
    const message = code === 'RATE_LIMITED'
      ? 'RATE_LIMITED'
      : code === 'PAYMENT_REQUIRED'
        ? 'PAYMENT_REQUIRED'
      : code === 'INVALID_CREDENTIAL'
        ? 'INVALID_CREDENTIAL'
        : payload.error?.message ?? 'MODEL_TEST_FAILED'
    throw new Error(message)
  }

  return payload.data
}

export async function getLLMStatus(): Promise<LLMConfigStatus> {
  const response = await fetch('/api/llm/status')
  const payload: { success: boolean; data?: LLMConfigStatus } = await response.json()
  if (!response.ok || !payload.success || !payload.data) {
    throw new Error('Unable to load model configuration.')
  }
  return payload.data
}

export async function interpretAnalysis(username: string, language: 'en' | 'fa'): Promise<InterpretationResult> {
  const response = await fetch('/api/interpret', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, language }),
  })
  const payload: { success: boolean; data?: InterpretationResult; error?: { code: string } } = await response.json()
  if (!response.ok || !payload.success || !payload.data) {
    throw new Error(payload.error?.code ?? 'LLM_PROVIDER_ERROR')
  }
  return payload.data
}
