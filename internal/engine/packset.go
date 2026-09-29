package engine

import (
	"slices"

	"github.com/jjmrocha/ai-toolkit/packs"
)

type packSet []packs.ToolPack

func (s *packSet) add(pack packs.ToolPack, err error) error {
	if err != nil {
		return err
	}

	*s = append(*s, pack)

	return nil
}

func (s *packSet) close() {
	for _, pack := range slices.Backward(*s) {
		_ = pack.Close()
	}
}
