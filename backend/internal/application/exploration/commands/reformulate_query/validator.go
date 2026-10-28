package reformulate_query

import (
	"errors"
	"strings"
)

func (c ReformulateQueryCommand) Validate() error {
	if strings.TrimSpace(c.Prompt) == "" {
		return errors.New("prompt cannot be empty")
	}
	return nil
}
