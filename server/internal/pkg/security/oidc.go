package security

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
)

const defaultOIDCSigningKeySize = 2048

type OIDCSigner struct {
	privateKey *rsa.PrivateKey
	keyID      string
}

func NewOIDCSigner(privateKeyPEM string, keyID string) (*OIDCSigner, error) {
	privateKey, err := loadOIDCPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, err
	}
	if privateKey == nil {
		privateKey, err = rsa.GenerateKey(rand.Reader, defaultOIDCSigningKeySize)
		if err != nil {
			return nil, fmt.Errorf("generate OIDC RSA key: %w", err)
		}
	}

	resolvedKeyID := strings.TrimSpace(keyID)
	if resolvedKeyID == "" {
		resolvedKeyID, err = deriveOIDCKeyID(&privateKey.PublicKey)
		if err != nil {
			return nil, err
		}
	}

	return &OIDCSigner{privateKey: privateKey, keyID: resolvedKeyID}, nil
}

func (signer *OIDCSigner) SignJWT(claims map[string]any) (string, error) {
	headerBytes, err := json.Marshal(map[string]string{
		"alg": "RS256",
		"kid": signer.keyID,
		"typ": "JWT",
	})
	if err != nil {
		return "", fmt.Errorf("marshal JWT header: %w", err)
	}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal JWT payload: %w", err)
	}

	signingInput := encodeJWTPart(headerBytes) + "." + encodeJWTPart(payloadBytes)
	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, signer.privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}

	return signingInput + "." + encodeJWTPart(signature), nil
}

func (signer *OIDCSigner) JWKS() map[string]any {
	publicKey := signer.privateKey.PublicKey
	return map[string]any{
		"keys": []map[string]any{{
			"alg": "RS256",
			"e":   encodeBigInt(big.NewInt(int64(publicKey.E))),
			"kid": signer.keyID,
			"kty": "RSA",
			"n":   encodeBigInt(publicKey.N),
			"use": "sig",
		}},
	}
}

func loadOIDCPrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	raw := strings.TrimSpace(strings.ReplaceAll(privateKeyPEM, `\n`, "\n"))
	if raw == "" {
		return nil, nil
	}

	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, fmt.Errorf("decode OIDC private key PEM: invalid PEM data")
	}

	switch block.Type {
	case "PRIVATE KEY":
		parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse PKCS8 OIDC private key: %w", err)
		}
		privateKey, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("parse PKCS8 OIDC private key: expected RSA private key")
		}
		return privateKey, nil
	case "RSA PRIVATE KEY":
		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse PKCS1 OIDC private key: %w", err)
		}
		return privateKey, nil
	default:
		return nil, fmt.Errorf("decode OIDC private key PEM: unsupported PEM type %s", block.Type)
	}
}

func deriveOIDCKeyID(publicKey *rsa.PublicKey) (string, error) {
	encodedKey, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("marshal OIDC public key: %w", err)
	}
	sum := sha256.Sum256(encodedKey)
	return base64.RawURLEncoding.EncodeToString(sum[:16]), nil
}

func encodeJWTPart(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}

func encodeBigInt(value *big.Int) string {
	return base64.RawURLEncoding.EncodeToString(value.Bytes())
}
