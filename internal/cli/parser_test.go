package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	sessionID  = "4th2w29y76ozr1zthqah9dbnm"
	profile    = "work"
	resumeFlag = "-resume"
)

func TestParse(t *testing.T) {
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
			result := Parse(tc.args)
			// then
			assert.Equal(t, &tc.expected, result)
		})
	}
}
