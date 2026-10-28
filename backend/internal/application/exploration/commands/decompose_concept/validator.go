package decompose_concept

import (
	"errors"
	"strings"
)

func (c DecomposeConceptCommand) Validate() error {
	trimmed := strings.TrimSpace(c.Prompt)
	if trimmed == "" {
		return errors.New("prompt cannot be empty")
	}
	if len(trimmed) < 3 {
		return errors.New("prompt must be at least 3 characters")
	}
	return nil
}
