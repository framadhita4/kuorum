package search_papers

import (
	"errors"
	"strings"
)

func (q SearchPapersQuery) Validate() error {
	if strings.TrimSpace(q.Query) == "" {
		return errors.New("search query cannot be empty")
	}
	if q.Limit < 0 {
		return errors.New("limit cannot be negative")
	}
	return nil
}
