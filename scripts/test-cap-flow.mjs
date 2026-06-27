import { createHash } from 'node:crypto'

function parseArgs(argv) {
	const args = {}
	for (let index = 0; index < argv.length; index += 1) {
		const current = argv[index]
		if (!current.startsWith('--')) {
			continue
		}
		const key = current.slice(2)
		const next = argv[index + 1]
		if (!next || next.startsWith('--')) {
			args[key] = 'true'
			continue
		}
		args[key] = next
		index += 1
	}
	return args
}

function usage() {
	console.log(`Usage:
  node scripts/test-cap-flow.mjs --cap-endpoint https://captcha.example.com --site-key abc123 --secret your-secret

Optional:
  --app-url https://account.example.com
  --email test@example.com

Examples:
  node scripts/test-cap-flow.mjs --cap-endpoint https://captcha.yealqp.cn --site-key 1cbf106b94 --secret xxx
  node scripts/test-cap-flow.mjs --cap-endpoint https://captcha.yealqp.cn --site-key 1cbf106b94 --secret xxx --app-url https://account.idcxl.cn --email you@example.com
`)
}

function normalizedEndpoint(base, siteKey) {
	return `${String(base || '').trim().replace(/\/+$/, '')}/${String(siteKey || '').trim().replace(/^\/+|\/+$/g, '')}/`
}

function deriveSeed(input, length) {
	let value = 2166136261
	for (let index = 0; index < input.length; index += 1) {
		value ^= input.charCodeAt(index)
		value += (value << 1) + (value << 4) + (value << 7) + (value << 8) + (value << 24)
	}
	value >>>= 0
	let output = ''
	while (output.length < length) {
		value ^= value << 13
		value ^= value >>> 17
		value ^= value << 5
		value >>>= 0
		output += value.toString(16).padStart(8, '0')
	}
	return output.slice(0, length)
}

function matchesTarget(hashHex, targetHex) {
	const totalBits = targetHex.length * 4
	const fullBytes = Math.floor(totalBits / 8)
	const partialBits = totalBits % 8
	const normalizedTarget = targetHex.length % 2 === 0 ? targetHex : `${targetHex}0`
	const target = Buffer.from(normalizedTarget, 'hex')
	const hash = Buffer.from(hashHex, 'hex')

	for (let index = 0; index < fullBytes; index += 1) {
		if (hash[index] !== target[index]) {
			return false
		}
	}

	if (partialBits > 0) {
		const mask = (255 << (8 - partialBits)) & 255
		if ((hash[fullBytes] & mask) !== (target[fullBytes] & mask)) {
			return false
		}
	}

	return true
}

function sha256Hex(value) {
	return createHash('sha256').update(value).digest('hex')
}

function solveChallenge(salt, target) {
	let nonce = 0
	for (;;) {
		if (matchesTarget(sha256Hex(`${salt}${nonce}`), target)) {
			return nonce
		}
		nonce += 1
	}
}

function expandChallenges(payload) {
	const challenge = payload.challenge
	if (Array.isArray(challenge)) {
		return challenge
	}
	let index = 0
	return Array.from({ length: challenge.c }, () => {
		index += 1
		return [
			deriveSeed(`${payload.token}${index}`, challenge.s),
			deriveSeed(`${payload.token}${index}d`, challenge.d),
		]
	})
}

async function requestJSON(url, init) {
	const response = await fetch(url, init)
	const text = await response.text()
	let data = null
	try {
		data = text ? JSON.parse(text) : null
	}
	catch {
		data = text
	}
	return { ok: response.ok, status: response.status, data }
}

async function main() {
	const args = parseArgs(process.argv.slice(2))
	const capEndpoint = args['cap-endpoint']
	const siteKey = args['site-key']
	const secret = args.secret
	const appURL = args['app-url']
	const email = args.email

	if (!capEndpoint || !siteKey || !secret) {
		usage()
		process.exitCode = 1
		return
	}

	const apiEndpoint = normalizedEndpoint(capEndpoint, siteKey)
	console.log('CAP endpoint:', apiEndpoint)

	console.log('\n[1/4] Requesting challenge...')
	const challengeResult = await requestJSON(`${apiEndpoint}challenge`, { method: 'POST' })
	console.log('challenge status:', challengeResult.status)
	console.log('challenge body:', JSON.stringify(challengeResult.data, null, 2))
	if (!challengeResult.ok || !challengeResult.data?.token || !challengeResult.data?.challenge) {
		process.exitCode = 1
		return
	}

	console.log('\n[2/4] Solving challenge...')
	const start = Date.now()
	const expanded = expandChallenges(challengeResult.data)
	const solutions = expanded.map(([salt, target], index) => {
		const nonce = solveChallenge(salt, target)
		console.log(`solved ${index + 1}/${expanded.length}: nonce=${nonce}`)
		return nonce
	})
	console.log(`solve completed in ${Date.now() - start}ms`)

	console.log('\n[3/4] Redeeming token...')
	const redeemResult = await requestJSON(`${apiEndpoint}redeem`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			token: challengeResult.data.token,
			solutions,
		}),
	})
	console.log('redeem status:', redeemResult.status)
	console.log('redeem body:', JSON.stringify(redeemResult.data, null, 2))
	if (!redeemResult.ok || !redeemResult.data?.success || !redeemResult.data?.token) {
		process.exitCode = 1
		return
	}

	console.log('\n[4/4] Verifying with siteverify...')
	const verifyResult = await requestJSON(`${apiEndpoint}siteverify`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			secret,
			response: redeemResult.data.token,
		}),
	})
	console.log('siteverify status:', verifyResult.status)
	console.log('siteverify body:', JSON.stringify(verifyResult.data, null, 2))

	if (appURL && email) {
		console.log('\n[extra] Calling app register code endpoint...')
		const registerResult = await requestJSON(`${String(appURL).trim().replace(/\/+$/, '')}/api/auth/register/code/send`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				email,
				captcha_token: redeemResult.data.token,
			}),
		})
		console.log('app endpoint status:', registerResult.status)
		console.log('app endpoint body:', JSON.stringify(registerResult.data, null, 2))
	}
}

main().catch(error => {
	console.error(error)
	process.exitCode = 1
})
