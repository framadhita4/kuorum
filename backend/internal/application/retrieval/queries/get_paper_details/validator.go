package get_paper_details

import (
	"errors"
	"strings"
)

func (q GetPaperDetailsQuery) Validate() error {
	if strings.TrimSpace(q.ID) == "" {
		return errors.New("paper id cannot be empty")
	}
	return nil
}
