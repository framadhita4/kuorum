package get_citations

import (
	"errors"
	"strings"
)

func (q GetCitationsQuery) Validate() error {
	if strings.TrimSpace(q.ID) == "" {
		return errors.New("paper id cannot be empty")
	}
	return nil
}
