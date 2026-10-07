package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sessionID  = "4th2w29y76ozr1zthqah9dbnm"
	profile    = "work"
	resumeFlag = "-resume"
	otherFlag  = "-verbose"
)

func TestParse(t *testing.T) {
	t.Run("reads the profile and the session to resume", func(t *testing.T) {
		cases := map[string]struct {
			args     []string
			expected Args
		}{
			"nothing":             {args: nil, expected: Args{Profile: "default"}},
			"profile only":        {args: []string{profile}, expected: Args{Profile: profile}},
			"resume only":         {args: []string{resumeFlag, sessionID}, expected: Args{Profile: "default", SessionID: sessionID}},
			"profile then resume": {args: []string{profile, resumeFlag, sessionID}, expected: Args{Profile: profile, SessionID: sessionID}},
		}

		for name, tc := range cases {
			t.Run(name, func(t *testing.T) {
				// when
				result, err := Parse(tc.args)
				// then
				require.NoError(t, err)
				assert.Equal(t, &tc.expected, result)
			})
		}
	})

	t.Run("asks for help", func(t *testing.T) {
		for _, flag := range []string{"-help", "--help", "-h"} {
			t.Run(flag, func(t *testing.T) {
				// when
				result, err := Parse([]string{flag})
				// then
				require.ErrorIs(t, err, ErrHelp)
				assert.Nil(t, result)
			})
		}
	})

	t.Run("rejects a malformed command line", func(t *testing.T) {
		cases := map[string][]string{
			"unknown flag":                {otherFlag},
			"resume without a session":    {resumeFlag},
			"profile then a session":      {profile, sessionID},
			"flag then resume":            {otherFlag, resumeFlag, sessionID},
			"profile then another flag":   {profile, otherFlag, sessionID},
			"more arguments than allowed": {profile, resumeFlag, sessionID, "extra"},
		}

		for name, args := range cases {
			t.Run(name, func(t *testing.T) {
				// when
				result, err := Parse(args)
				// then
				require.ErrorIs(t, err, ErrUsage)
				assert.Nil(t, result)
			})
		}
	})
}

func TestUsage(t *testing.T) {
	t.Run("writes the usage text", func(t *testing.T) {
		// given
		var out strings.Builder
		expected := "Usage: joe [profile] [-resume <id>]\n" +
			"\n" +
			"       profile   string   name of the profile (defaults to 'default')\n" +
			"       -resume   string   id of an exported session to resume\n" +
			"       -help              this message\n"
		// when
		Usage(&out)
		// then
		assert.Equal(t, expected, out.String())
	})
}
