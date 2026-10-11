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

export async function interpretAnalysis(username: string, language: 'en' | 'fa', signal?: AbortSignal): Promise<InterpretationResult> {
  const response = await fetch('/api/interpret', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, language }),
    signal,
  })
  const payload: { success: boolean; data?: InterpretationResult; error?: { code: string } } = await response.json()
  if (!response.ok || !payload.success || !payload.data) {
    throw new Error(payload.error?.code ?? 'LLM_PROVIDER_ERROR')
  }
  return payload.data
}

export async function streamInterpretAnalysis(
  username: string,
  language: 'en' | 'fa',
  onReport: (report: string) => void,
  signal?: AbortSignal,
): Promise<InterpretationResult> {
  const query = new URLSearchParams({ username, language })
  const response = await fetch(`/api/interpret/stream?${query}`, { signal })
  if (!response.ok) {
    const payload: { error?: { code?: string } } = await response.clone().json().catch(() => ({}))
    if ((response.status === 404 && !payload.error?.code) || response.status === 405) {
      return interpretAnalysis(username, language, signal)
    }
    throw new Error(payload.error?.code ?? 'LLM_PROVIDER_ERROR')
  }
  if (!response.headers.get('content-type')?.includes('text/event-stream')) {
    return interpretAnalysis(username, language, signal)
  }

  let buffer = ''
  let report = ''
  const completion: { result?: InterpretationResult } = {}
  const dispatch = (block: string) => {
    let event = 'message'
    const data: string[] = []
    for (const line of block.replace(/\r\n/g, '\n').replace(/\r/g, '\n').split('\n')) {
      if (line.startsWith(':')) continue
      if (line.startsWith('event:')) event = line.slice(6).trim()
      else if (line.startsWith('data:')) data.push(line.slice(5).trimStart())
    }
    if (!data.length) return
    const payload: unknown = JSON.parse(data.join('\n'))
    if (event === 'token') {
      const token = (payload as { text?: unknown }).text
      if (typeof token === 'string') {
        report += token
        onReport(report)
      }
    } else if (event === 'done') {
      completion.result = payload as InterpretationResult
    } else if (event === 'error') {
      const code = (payload as { code?: unknown }).code
      throw new Error(typeof code === 'string' ? code : 'LLM_PROVIDER_ERROR')
    }
  }
  const consume = (chunk: string, final = false) => {
    buffer += chunk
    while (true) {
      const boundary = /\r?\n\r?\n/.exec(buffer)
      if (!boundary || boundary.index === undefined) break
      dispatch(buffer.slice(0, boundary.index))
      buffer = buffer.slice(boundary.index + boundary[0].length)
    }
    if (final && buffer.trim()) {
      dispatch(buffer)
      buffer = ''
    }
  }

  if (!response.body) {
    return interpretAnalysis(username, language, signal)
  } else {
    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    while (true) {
      const { done, value } = await reader.read()
      if (done) {
        consume(decoder.decode(), true)
        break
      }
      consume(decoder.decode(value, { stream: true }))
    }
  }
  if (!completion.result) throw new Error('LLM_STREAM_INCOMPLETE')
  return { ...completion.result, report: completion.result.report || report }
}
