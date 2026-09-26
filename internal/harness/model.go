package harness

type Paths struct {
	ConfigDir string
	Home      string
	Repo      string
}

type Block struct {
	Path    string
	Content string
}

type Harness struct {
	Kind   Kind
	Blocks []Block
}
