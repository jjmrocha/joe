package harness

import "errors"

var ErrInvalidKind = errors.New("harness is not claude or agents")
