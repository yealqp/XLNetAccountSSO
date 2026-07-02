/**
 * Cap.js PoW 验证码求解器
 *
 * SHA-256 碰撞求解在 Web Worker 中执行，完全不阻塞主线程。
 */

// ── HTTP ──

interface ChallengeResponse {
  token: string
  count: number
  saltLength: number
  difficulty: number
}

async function fetchChallenge(apiEndpoint: string): Promise<ChallengeResponse> {
  const res = await fetch(`${apiEndpoint}challenge`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: '{}',
  })
  if (!res.ok) throw new Error(`获取挑战失败 (HTTP ${res.status})`)
  const body = await res.json()
  const root = (body.data || body) ?? {}
  const token = root.token as string | undefined
  const challenge = root.challenge as Record<string, unknown> | undefined
  if (!token || !challenge) throw new Error('挑战接口缺少 token/challenge')

  const count = Number(challenge.c)
  const saltLength = Number(challenge.s)
  const difficulty = Number(challenge.d)
  if (!Number.isFinite(count) || !Number.isFinite(saltLength) || !Number.isFinite(difficulty)) {
    throw new Error('挑战参数解析失败')
  }

  return { token, count, saltLength, difficulty }
}

async function redeemSolution(
  apiEndpoint: string,
  token: string,
  solutions: number[],
): Promise<string> {
  const res = await fetch(`${apiEndpoint}redeem`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token, solutions }),
  })
  if (!res.ok) throw new Error(`提交解答失败 (HTTP ${res.status})`)
  const body = await res.json()
  const root = (body.data || body) as Record<string, unknown>
  if (root.success === false) throw new Error(`服务端拒绝: ${String(root.message || 'unknown')}`)
  if (!root.token) throw new Error('服务端未返回 token')
  return String(root.token)
}

// ── Worker 求解 ──

let worker: Worker | null = null

function getWorker(): Worker {
  if (!worker) {
    worker = new Worker(new URL('./captcha.worker.ts', import.meta.url), { type: 'module' })
  }
  return worker
}

export function destroyWorker(): void {
  worker?.terminate()
  worker = null
}

// ── 公开 API ──

export async function solveCaptcha(
  apiEndpoint: string,
  onProgress?: (pct: number) => void,
): Promise<string> {
  onProgress?.(5)

  const challenge = await fetchChallenge(apiEndpoint)
  onProgress?.(15)

  const { token, count, saltLength, difficulty } = challenge

  // 在 Worker 中求解 PoW，主线程不阻塞
  const solutions = await new Promise<number[]>((resolve, reject) => {
    const w = getWorker()

    const onMessage = (e: MessageEvent<{ type: string; pct?: number; solutions?: number[]; error?: string }>) => {
      if (e.data.type === 'progress') {
        onProgress?.(15 + Math.round(((e.data.pct ?? 0) / 100) * 55))
      } else if (e.data.type === 'done') {
        w.removeEventListener('message', onMessage)
        w.removeEventListener('error', onError)
        resolve(e.data.solutions ?? [])
      }
    }

    const onError = (err: ErrorEvent) => {
      w.removeEventListener('message', onMessage)
      w.removeEventListener('error', onError)
      reject(new Error(err.message || 'Worker 求解失败'))
    }

    w.addEventListener('message', onMessage)
    w.addEventListener('error', onError)

    w.postMessage({ type: 'solve', token, count, saltLength, difficulty })
  })

  if (solutions.length === 0) throw new Error('求解器未产生任何 nonce')
  onProgress?.(70)

  const resultToken = await redeemSolution(apiEndpoint, token, solutions)
  onProgress?.(100)

  return resultToken
}
