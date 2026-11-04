package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sessionID      = "4th2w29y76ozr1zthqah9dbnm"
	defaultProfile = "default"
	profile        = "work"
	resumeFlag     = "-resume"
	debugFlag      = "-debug"
	otherFlag      = "-verbose"
)

func TestParse(t *testing.T) {
	t.Run("reads the profile, the session to resume and the debug mode", func(t *testing.T) {
		cases := map[string]struct {
			args     []string
			expected Args
		}{
			"nothing":             {args: nil, expected: Args{Profile: defaultProfile}},
			"profile only":        {args: []string{profile}, expected: Args{Profile: profile}},
			"resume only":         {args: []string{resumeFlag, sessionID}, expected: Args{Profile: defaultProfile, SessionID: sessionID}},
			"profile then resume": {args: []string{profile, resumeFlag, sessionID}, expected: Args{Profile: profile, SessionID: sessionID}},
			"debug only":          {args: []string{debugFlag}, expected: Args{Profile: defaultProfile, Debug: true}},
			"profile then debug":  {args: []string{profile, debugFlag}, expected: Args{Profile: profile, Debug: true}},
			"debug then resume":   {args: []string{debugFlag, resumeFlag, sessionID}, expected: Args{Profile: defaultProfile, SessionID: sessionID, Debug: true}},
			"resume then debug":   {args: []string{resumeFlag, sessionID, debugFlag}, expected: Args{Profile: defaultProfile, SessionID: sessionID, Debug: true}},
			"profile, resume and debug": {
				args:     []string{profile, resumeFlag, sessionID, debugFlag},
				expected: Args{Profile: profile, SessionID: sessionID, Debug: true},
			},
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
			"debug twice":                 {debugFlag, debugFlag},
			"debug then resume alone":     {debugFlag, resumeFlag},
			"debug between resume and id": {resumeFlag, debugFlag, sessionID},
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
		expected := "Usage: joe [profile] [-resume <id>] [-debug]\n" +
			"\n" +
			"       profile   string   name of the profile (defaults to 'default')\n" +
			"       -resume   string   id of an exported session to resume\n" +
			"       -debug             save each subagent run to ~/.config/joe/debug\n" +
			"       -help              this message\n"
		// when
		Usage(&out)
		// then
		assert.Equal(t, expected, out.String())
	})
}
