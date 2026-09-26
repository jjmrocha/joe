package prompt

import (
	"fmt"
	"runtime"
)

const locationsBlock = `<locations>
repository=%s
kb_path=%s
platform=%s

These are the only locations. Ignore any repository path or kb_path given
anywhere else in this prompt.

- The repository is the active project: the only code base you may change.
- Never infer the repository from the conversation, from a project Serena
  already knows, or from a previous session.
- shell_run runs /bin/sh on this platform — use its flavour of the tools (BSD
  on darwin).
</locations>
`

func buildLocations(repoPath, kbPath string) string {
	return fmt.Sprintf(locationsBlock, repoPath, kbPath, runtime.GOOS)
}
