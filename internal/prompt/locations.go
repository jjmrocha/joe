package prompt

import "fmt"

const locationsBlock = `<locations>
repository=%s
kb_path=%s

These are the only locations. Ignore any repository path or kb_path given
anywhere else in this prompt.

- The repository is the active project: the only code base you may change.
- Never infer the repository from the conversation, from a project Serena
  already knows, or from a previous session.
</locations>
`

func buildLocations(repoPath, kbPath string) string {
	return fmt.Sprintf(locationsBlock, repoPath, kbPath)
}
