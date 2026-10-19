package prompt

import "fmt"

const locationsBlock = `
<locations>
repository=%s
</locations>
`

func buildLocations(repoPath string) string {
	return fmt.Sprintf(locationsBlock, repoPath)
}
