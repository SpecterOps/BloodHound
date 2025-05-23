package validation

import (
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/src/utils"
	"github.com/teambition/rrule-go"
)

const (
	ErrInvalidRrule = "invalid rrule specified: %s"
)

// type RRuleValidator struct implements the Validator interface which allows the usage of the `rrule` struct tag
// to ensure a string is in a valid rrule format by calling `Validate`
type RRuleValidator struct {
}

// NewRRuleValidator returns a new Validator
func NewRRuleValidator(_ map[string]string) Validator {
	return RRuleValidator{}
}

func (s RRuleValidator) Validate(value any) utils.Errors {
	var (
		rruleStr string
		ok       bool
		errs     = utils.Errors{}
	)

	if rruleStr, ok = value.(string); !ok {
		return append(errs, fmt.Errorf(ErrInvalidRrule, value))
	}

	//Validate that the rrule is a good rule. We're going to require a DTSTART to keep scheduling consistent.
	//We're also going to reject UNTIL/COUNT because it will most likely break the pipeline once it's hit without being invalid
	if rruleStr == "" {
		return nil
	} else if _, err := rrule.StrToRRule(rruleStr); err != nil {
		return append(errs, fmt.Errorf(ErrInvalidRrule, err))
	} else if strings.Contains(strings.ToUpper(rruleStr), "UNTIL") {
		return append(errs, fmt.Errorf(ErrInvalidRrule, "until not supported"))
	} else if strings.Contains(strings.ToUpper(rruleStr), "COUNT") {
		return append(errs, fmt.Errorf(ErrInvalidRrule, "count not supported"))
	} else if !strings.Contains(strings.ToUpper(rruleStr), "DTSTART") {
		return append(errs, fmt.Errorf(ErrInvalidRrule, "dtstart is required"))
	}

	return nil
}
