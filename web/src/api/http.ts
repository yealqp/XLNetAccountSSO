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

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers)
  const isFormBody = options.body instanceof FormData || options.body instanceof URLSearchParams

  if (options.body && !isFormBody && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(path, {
    credentials: 'include',
    ...options,
    headers,
  })

  if (response.status === 204) {
    return undefined as T
  }

  const rawText = await response.text()
  const data = rawText ? JSON.parse(rawText) : null

  if (!response.ok) {
    const message = typeof data?.message === 'string' ? data.message : 'Request failed'
    throw new ApiError(message, response.status, data)
  }

  return data as T
}
