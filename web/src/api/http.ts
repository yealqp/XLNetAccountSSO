import { getAuthToken } from '@/utils/authToken'
import { handleUnauthorizedRedirect } from '@/utils/authExpiry'

export class ApiError extends Error {
  status: number
  data: unknown

  constructor(message: string, status: number, data: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.data = data
  }
}

interface ApiEnvelope<T> {
  code: number
  data: T
  message: string
}

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers)
  const isFormBody = options.body instanceof FormData || options.body instanceof URLSearchParams

  if (options.body && !isFormBody && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const authToken = getAuthToken()
  if (authToken && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${authToken}`)
  }

  const response = await fetch(path, {
    ...options,
    headers,
  })

  if (response.status === 204) {
    return undefined as T
  }

  const rawText = await response.text()
  const data = parseResponseData(rawText)

  const envelope = isApiEnvelope<T>(data) ? data : null

  if (!response.ok) {
    const message = envelope?.message
      || (typeof data === 'object' && data !== null && 'message' in data && typeof data.message === 'string'
        ? data.message
        : typeof data === 'string' && data.trim()
          ? data
          : 'Request failed')
    if (response.status === 401) {
      void handleUnauthorizedRedirect()
    }
    throw new ApiError(message, response.status, envelope?.data ?? data)
  }

  if (envelope) {
    return envelope.data as T
  }

  return data as T
}

function parseResponseData(rawText: string) {
  if (!rawText) {
    return null
  }

  try {
    return JSON.parse(rawText) as unknown
  }
  catch {
    return rawText
  }
}

function isApiEnvelope<T>(value: unknown): value is ApiEnvelope<T> {
  return Boolean(
    value
    && typeof value === 'object'
    && 'code' in (value as Record<string, unknown>)
    && 'data' in (value as Record<string, unknown>)
    && 'message' in (value as Record<string, unknown>),
  )
}
