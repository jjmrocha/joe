package session

import "errors"

var (
	ErrInvalidSessionID = errors.New("invalid session id")
	ErrSessionNotFound  = errors.New("session not found")
	ErrRepoMismatch     = errors.New("session belongs to another repository")
	ErrUnknownRole      = errors.New("unknown role")
)
