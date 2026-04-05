package service

import "errors"

var (
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidInput   = errors.New("invalid input")
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")
	ErrInvalidGrant   = errors.New("invalid grant")
	ErrAccessDenied   = errors.New("access denied")
	ErrInvalidClient  = errors.New("invalid client")
	ErrInvalidToken   = errors.New("invalid token")
	ErrInvalidSession = errors.New("invalid session")
)
