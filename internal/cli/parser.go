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

	for len(args) > 0 {
		switch {
		case args[0] == "-debug" && !result.Debug:
			result.Debug = true
			args = args[1:]
		case args[0] == "-resume" && len(args) > 1 && result.SessionID == "":
			result.SessionID = args[1]
			args = args[2:]
		default:
			return nil, ErrUsage
		}
	}

	return &result, nil
}

func Usage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage: joe [profile] [-resume <id>] [-debug]")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "       profile   string   name of the profile (defaults to 'default')")
	_, _ = fmt.Fprintln(w, "       -resume   string   id of an exported session to resume")
	_, _ = fmt.Fprintln(w, "       -debug             save each subagent run to ~/.config/joe/debug")
	_, _ = fmt.Fprintln(w, "       -help              this message")
}
