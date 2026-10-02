package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/jjmrocha/joe/internal/config"
)

func Parse(args []string) *Args {
	result := Args{Profile: config.DefaultProfile}

	switch len(args) {
	case 0:
		return &result
	case 1:
		switch args[0] {
		case "-help", "--help", "-h":
			help(0)
		default:
			if strings.HasPrefix(args[0], "-") {
				help(1)
			}

			result.Profile = args[0]
		}
	case 2:
		if args[0] == "-resume" {
			result.SessionID = args[1]
		} else {
			help(1)
		}
	case 3:
		if strings.HasPrefix(args[0], "-") || args[1] != "-resume" {
			help(1)
		}

		result.Profile = args[0]
		result.SessionID = args[2]
	default:
		help(1)
	}

	return &result
}

func help(code int) {
	fmt.Println("Usage: joe [profile] [-resume <id>]")
	fmt.Println()
	fmt.Println("       profile   string   name of the profile (defaults to 'default')")
	fmt.Println("       -resume   string   id of an exported session to resume")
	fmt.Println("       -help              this message")
	os.Exit(code)
}
