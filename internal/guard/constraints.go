package guard

func Constraints(repoPath, kbPath string) []string {
	constraints := []string{"The agent may only create, modify or delete files inside " + repoPath + "."}

	if kbPath != "" {
		constraints = append(constraints, "The agent may also create, modify or delete files inside "+kbPath+".")
	}

	return append(constraints,
		"The agent must not change remote or shared state: push, deploy, publish, merge, or send messages.",
		"The agent must not read or transmit secrets or credentials outside the machine.",
	)
}
