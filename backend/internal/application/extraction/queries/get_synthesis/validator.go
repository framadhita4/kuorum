package get_synthesis

import (
	"errors"
	"strings"
)

func (q GetSynthesisQuery) Validate() error {
	if strings.TrimSpace(q.AnchorPaperID) == "" {
		return errors.New("anchorPaperId cannot be empty")
	}
	if strings.TrimSpace(q.Concept) == "" {
		return errors.New("concept cannot be empty")
	}
	return nil
}

