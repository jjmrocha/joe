package cli

import "errors"

var (
	ErrHelp  = errors.New("help requested")
	ErrUsage = errors.New("invalid command line")
)
