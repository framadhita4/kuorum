package get_graph

import (
	"errors"
	"strings"
)

func (q GetCitationGraphQuery) Validate() error {
	if strings.TrimSpace(q.AnchorPaperID) == "" {
		return errors.New("anchorPaperId cannot be empty")
	}
	return nil
}

