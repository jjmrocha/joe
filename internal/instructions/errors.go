package instructions

import "errors"

var ErrInvalidKind = errors.New("instructions is not claude or agents")
