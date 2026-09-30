package guard

import "errors"

var (
	ErrToolCallRejected = errors.New("rejected by joe: the call violates the session constraints")
	ErrToolCallTooLarge = errors.New("rejected by joe: the request is too big")
)
