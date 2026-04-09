export type Base64UrlString = string

export interface WebAuthnCredentialDescriptorJSON {
	type: PublicKeyCredentialType
	id: Base64UrlString
	transports?: AuthenticatorTransport[]
}

export interface PublicKeyCredentialUserEntityJSON extends Omit<PublicKeyCredentialUserEntity, 'id'> {
	id: Base64UrlString
}

export interface PublicKeyCredentialCreationOptionsJSON {
	rp: PublicKeyCredentialRpEntity
	user: PublicKeyCredentialUserEntityJSON
	challenge: Base64UrlString
	pubKeyCredParams: PublicKeyCredentialParameters[]
	timeout?: number
	excludeCredentials?: WebAuthnCredentialDescriptorJSON[]
	authenticatorSelection?: AuthenticatorSelectionCriteria
	attestation?: AttestationConveyancePreference
	extensions?: AuthenticationExtensionsClientInputs
}

export interface PublicKeyCredentialRequestOptionsJSON {
	challenge: Base64UrlString
	timeout?: number
	rpId?: string
	allowCredentials?: WebAuthnCredentialDescriptorJSON[]
	userVerification?: UserVerificationRequirement
	extensions?: AuthenticationExtensionsClientInputs
}

export interface PublicKeyCredentialCreationOptionsEnvelopeJSON {
	publicKey: PublicKeyCredentialCreationOptionsJSON
	mediation?: CredentialMediationRequirement
}

export interface PublicKeyCredentialRequestOptionsEnvelopeJSON {
	publicKey: PublicKeyCredentialRequestOptionsJSON
	mediation?: CredentialMediationRequirement
}

export interface RegistrationCredentialJSON {
	id: string
	rawId: Base64UrlString
	type: PublicKeyCredentialType
	authenticatorAttachment?: AuthenticatorAttachment | null
	clientExtensionResults: AuthenticationExtensionsClientOutputs
	response: {
		clientDataJSON: Base64UrlString
		attestationObject: Base64UrlString
		transports?: AuthenticatorTransport[]
		publicKeyAlgorithm?: number
		publicKey?: Base64UrlString | null
	}
}

export interface AuthenticationCredentialJSON {
	id: string
	rawId: Base64UrlString
	type: PublicKeyCredentialType
	authenticatorAttachment?: AuthenticatorAttachment | null
	clientExtensionResults: AuthenticationExtensionsClientOutputs
	response: {
		clientDataJSON: Base64UrlString
		authenticatorData: Base64UrlString
		signature: Base64UrlString
		userHandle?: Base64UrlString | null
	}
}
