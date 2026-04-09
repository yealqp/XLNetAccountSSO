import type {
	AuthenticationCredentialJSON,
	Base64UrlString,
	PublicKeyCredentialCreationOptionsEnvelopeJSON,
	PublicKeyCredentialRequestOptionsEnvelopeJSON,
	RegistrationCredentialJSON,
	WebAuthnCredentialDescriptorJSON,
} from '@/types/webauthn'

export function getPasskeySupportMessage() {
	if (typeof window === 'undefined' || typeof navigator === 'undefined') {
		return '当前环境不支持通行密钥。'
	}

	if (!window.isSecureContext) {
		return '当前页面未启用安全上下文，请使用 HTTPS 或 localhost 后再试。'
	}

	if (typeof window.PublicKeyCredential === 'undefined') {
		return '当前浏览器不支持通行密钥。'
	}

	if (!navigator.credentials || typeof navigator.credentials.create !== 'function' || typeof navigator.credentials.get !== 'function') {
		return '当前浏览器缺少通行密钥能力，请升级浏览器后再试。'
	}

	return null
	}

export async function createPasskeyCredential(options: PublicKeyCredentialCreationOptionsEnvelopeJSON) {
	assertPasskeySupport()

	const credential = await navigator.credentials.create({
		publicKey: toCreationOptions(options.publicKey),
	})

	if (!(credential instanceof PublicKeyCredential)) {
		throw new Error('浏览器未返回有效的通行密钥注册结果。')
	}

	if (!(credential.response instanceof AuthenticatorAttestationResponse)) {
		throw new Error('浏览器未返回有效的注册响应。')
	}

	return serializeRegistrationCredential(credential)
}

export async function getPasskeyCredential(options: PublicKeyCredentialRequestOptionsEnvelopeJSON) {
	assertPasskeySupport()

	const credential = await navigator.credentials.get({
		publicKey: toRequestOptions(options.publicKey),
		mediation: options.mediation,
	})

	if (!(credential instanceof PublicKeyCredential)) {
		throw new Error('浏览器未返回有效的通行密钥登录结果。')
	}

	if (!(credential.response instanceof AuthenticatorAssertionResponse)) {
		throw new Error('浏览器未返回有效的登录响应。')
	}

	return serializeAuthenticationCredential(credential)
}

export function describePasskeyError(error: unknown) {
	if (error instanceof DOMException) {
		switch (error.name) {
			case 'AbortError':
				return '通行密钥操作已中止，请重试。'
			case 'ConstraintError':
				return '当前设备无法满足通行密钥创建条件。'
			case 'InvalidStateError':
				return '该通行密钥已存在，或当前状态不允许继续。'
			case 'NotAllowedError':
				return '您已取消通行密钥操作，或操作已超时。'
			case 'NotSupportedError':
				return '当前浏览器不支持通行密钥。'
			case 'SecurityError':
				return '当前页面未满足通行密钥安全要求，请确认使用 HTTPS 或 localhost。'
			default:
				return error.message || '通行密钥操作失败，请稍后重试。'
		}
	}

	if (error instanceof Error) {
		return error.message || '通行密钥操作失败，请稍后重试。'
	}

	return '通行密钥操作失败，请稍后重试。'
	}

function assertPasskeySupport() {
	const supportMessage = getPasskeySupportMessage()

	if (supportMessage) {
		throw new Error(supportMessage)
	}
}

function toCreationOptions(options: PublicKeyCredentialCreationOptionsEnvelopeJSON['publicKey']): PublicKeyCredentialCreationOptions {
	return {
		...options,
		challenge: decodeBase64Url(options.challenge),
		user: {
			...options.user,
			id: decodeBase64Url(options.user.id),
		},
		excludeCredentials: options.excludeCredentials?.map(toCredentialDescriptor),
	}
}

function toRequestOptions(options: PublicKeyCredentialRequestOptionsEnvelopeJSON['publicKey']): PublicKeyCredentialRequestOptions {
	return {
		...options,
		challenge: decodeBase64Url(options.challenge),
		allowCredentials: options.allowCredentials?.map(toCredentialDescriptor),
	}
}

function toCredentialDescriptor(descriptor: WebAuthnCredentialDescriptorJSON): PublicKeyCredentialDescriptor {
	return {
		...descriptor,
		id: decodeBase64Url(descriptor.id),
	}
}

function serializeRegistrationCredential(credential: PublicKeyCredential): RegistrationCredentialJSON {
	const response = credential.response as AuthenticatorAttestationResponse & {
		getPublicKey?: () => ArrayBuffer | null
		getPublicKeyAlgorithm?: () => number
		getTransports?: () => AuthenticatorTransport[]
	}

	return {
		id: credential.id,
		rawId: encodeBase64Url(credential.rawId),
		type: 'public-key',
		authenticatorAttachment: normalizeAuthenticatorAttachment(credential.authenticatorAttachment),
		clientExtensionResults: credential.getClientExtensionResults(),
		response: {
			clientDataJSON: encodeBase64Url(response.clientDataJSON),
			attestationObject: encodeBase64Url(response.attestationObject),
			transports: typeof response.getTransports === 'function' ? normalizeAuthenticatorTransports(response.getTransports()) : undefined,
			publicKeyAlgorithm: typeof response.getPublicKeyAlgorithm === 'function' ? response.getPublicKeyAlgorithm() : undefined,
			publicKey: typeof response.getPublicKey === 'function'
				? encodeOptionalBase64Url(response.getPublicKey())
				: undefined,
		},
	}
}

function serializeAuthenticationCredential(credential: PublicKeyCredential): AuthenticationCredentialJSON {
	const response = credential.response as AuthenticatorAssertionResponse

	return {
		id: credential.id,
		rawId: encodeBase64Url(credential.rawId),
		type: 'public-key',
		authenticatorAttachment: normalizeAuthenticatorAttachment(credential.authenticatorAttachment),
		clientExtensionResults: credential.getClientExtensionResults(),
		response: {
			clientDataJSON: encodeBase64Url(response.clientDataJSON),
			authenticatorData: encodeBase64Url(response.authenticatorData),
			signature: encodeBase64Url(response.signature),
			userHandle: response.userHandle ? encodeBase64Url(response.userHandle) : null,
		},
	}
}

function decodeBase64Url(input: Base64UrlString) {
	const base64 = input.replace(/-/g, '+').replace(/_/g, '/')
	const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '=')
	const binary = window.atob(padded)
	const bytes = new Uint8Array(binary.length)

	for (let index = 0; index < binary.length; index += 1) {
		bytes[index] = binary.charCodeAt(index)
	}

	return bytes.buffer
}

function encodeBase64Url(value: ArrayBuffer | ArrayBufferView) {
	const bytes = value instanceof ArrayBuffer
		? new Uint8Array(value)
		: new Uint8Array(value.buffer, value.byteOffset, value.byteLength)

	let binary = ''
	const chunkSize = 0x8000

	for (let index = 0; index < bytes.length; index += chunkSize) {
		binary += String.fromCharCode(...bytes.subarray(index, index + chunkSize))
	}

	return window.btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
}

function encodeOptionalBase64Url(value: ArrayBuffer | null) {
	if (!value) {
		return null
	}

	return encodeBase64Url(value)
}

function normalizeAuthenticatorAttachment(value: string | null): AuthenticatorAttachment | null {
	return value === 'cross-platform' || value === 'platform' ? value : null
}

function normalizeAuthenticatorTransports(values: string[]): AuthenticatorTransport[] {
	return values.filter((value): value is AuthenticatorTransport => {
		return value === 'ble'
			|| value === 'hybrid'
			|| value === 'internal'
			|| value === 'nfc'
			|| value === 'smart-card'
			|| value === 'usb'
	})
}
