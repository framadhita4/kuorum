package get_references

import (
	"errors"
	"strings"
)

func (q GetReferencesQuery) Validate() error {
	if strings.TrimSpace(q.ID) == "" {
		return errors.New("paper id cannot be empty")
	}
	return nil
}
