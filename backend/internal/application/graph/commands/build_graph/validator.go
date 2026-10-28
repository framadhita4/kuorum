package build_graph

import (
	"errors"
	"strings"
)

func (c BuildCitationGraphCommand) Validate() error {
	if strings.TrimSpace(c.AnchorPaperID) == "" {
		return errors.New("anchorPaperId cannot be empty")
	}
	if c.MaxReferences < 0 || c.MaxReferences > 50 {
		return errors.New("maxReferences must be between 0 and 50")
	}
	if c.MaxCitations < 0 || c.MaxCitations > 50 {
		return errors.New("maxCitations must be between 0 and 50")
	}
	return nil
}

