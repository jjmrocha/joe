package prompt

import (
	"fmt"

	"github.com/jjmrocha/joe/internal/guard"
)

const guardBlock = `
<guard>
Every tool call you make is screened against the file, remote-state and secret
rules in <boundaries> before it runs, and one that breaks them is refused with
%q.

When a call is refused, it is the change that was judged out of bounds — not
the tool. Do not re-attempt the same change through another tool, and do not
argue with the refusal: it is joe's gate, not the classifier's advice. Stop,
tell the user what was refused and why, and let them decide.
</guard>
`

func buildGuard() string {
	return fmt.Sprintf(guardBlock, guard.ErrToolCallRejected.Error())
}
