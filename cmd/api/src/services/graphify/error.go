package graphify

import (
	"errors"
	"strings"
)

func newGraphifyErrorBuilder() graphifyErrorBuilder {
	return graphifyErrorBuilder{}
}

type graphifyErrorBuilder struct {
	Errors []error
}

func (s *graphifyErrorBuilder) Add(e error) {
	var graphifyError GraphifyError
	if ok := errors.As(e, &graphifyError); ok {
		s.Errors = append(s.Errors, graphifyError.Errors...)
	} else if e != nil {
		s.Errors = append(s.Errors, e)
	}
}

func (s graphifyErrorBuilder) Build() error {
	if len(s.Errors) == 0 {
		return nil
	} else {
		return GraphifyError{Errors: s.Errors}
	}
}

type GraphifyError struct {
	Errors []error
}

func (s GraphifyError) AsStrings() []string {
	errStrings := make([]string, len(s.Errors))

	for i, err := range s.Errors {
		errStrings[i] = err.Error()
	}

	return errStrings
}

func (s GraphifyError) Error() string {
	return strings.Join(s.AsStrings(), "; ")
}
