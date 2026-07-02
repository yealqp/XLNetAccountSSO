import { ApiError } from '@/api/http'
import { Message } from '@arco-design/web-vue'

export function extractErrorMessage(error: unknown, fallback: string): string {
  return error instanceof ApiError ? error.message : fallback
}

export function handleApiError(error: unknown, fallback: string): void {
  const msg = extractErrorMessage(error, fallback)
  Message.error(msg)
}
