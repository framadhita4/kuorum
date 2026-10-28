package synthesize_concept

import (
	"errors"
	"strings"
)

func (c SynthesizeConceptCommand) Validate() error {
	if strings.TrimSpace(c.AnchorPaperID) == "" {
		return errors.New("anchorPaperId cannot be empty")
	}
	if strings.TrimSpace(c.Concept) == "" {
		return errors.New("concept cannot be empty")
	}
	return nil
}

