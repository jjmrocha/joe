package harness

import "errors"

var (
	ErrInvalidKBPath = errors.New("kb_path is not absolute")
	ErrInvalidKind   = errors.New("harness is not claude or agents")
)
