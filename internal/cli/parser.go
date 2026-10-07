package cli

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/jjmrocha/joe/internal/config"
)

func Parse(args []string) (*Args, error) {
	result := Args{Profile: config.DefaultProfile}

	if len(args) == 1 && slices.Contains([]string{"-help", "--help", "-h"}, args[0]) {
		return nil, ErrHelp
	}

	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		result.Profile = args[0]
		args = args[1:]
	}

	switch len(args) {
	case 0:
		return &result, nil
	case 2:
		if args[0] == "-resume" {
			result.SessionID = args[1]
			return &result, nil
		}
	}

	return nil, ErrUsage
}

func Usage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage: joe [profile] [-resume <id>]")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "       profile   string   name of the profile (defaults to 'default')")
	_, _ = fmt.Fprintln(w, "       -resume   string   id of an exported session to resume")
	_, _ = fmt.Fprintln(w, "       -help              this message")
}
