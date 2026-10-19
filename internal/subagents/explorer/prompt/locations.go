package prompt

import "fmt"

const locationsBlock = `
<locations>
repository=%s
</locations>
`

const locationsWithKBBlock = `
<locations>
repository=%s
kb_path=%s
</locations>
`

func buildLocations(repoPath, kbPath string) string {
	if kbPath == "" {
		return fmt.Sprintf(locationsBlock, repoPath)
	}

	return fmt.Sprintf(locationsWithKBBlock, repoPath, kbPath)
}
