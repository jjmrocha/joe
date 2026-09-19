package harness

type Paths struct {
	ConfigDir string
	Home      string
	Repo      string
}

type Harness struct {
	Kind   Kind
	Blocks []string
}
