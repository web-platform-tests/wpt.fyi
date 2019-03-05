package shared

import (
	"fmt"
	"strings"
)

type MultiError struct {
	errors []error
	errStr string
	when   string
}

func NewMultiErrorFromChan(errors chan error, when string) *MultiError {
	var multiError MultiError
	for err := range errors {
		multiError.errors = append(multiError.errors, err)
		multiError.errStr += strings.TrimSpace(err.Error()) + "\n"
	}
	if multiError.errors != nil {
		multiError.when = when
		return &multiError
	}
	return nil
}

func (e MultiError) Error() string {
	if len(e.errors) == 0 {
		return ""
	}
	return fmt.Sprintf("%d error(s) occured when %s:\n%s", len(e.errors), e.when, e.errStr)
}

func (e MultiError) Count() int {
	return len(e.errors)
}
