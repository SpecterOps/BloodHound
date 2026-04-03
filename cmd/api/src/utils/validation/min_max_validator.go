package validation

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/specterops/bloodhound/cmd/api/src/utils"
)

const (
	ErrorNonInt = "non integer provided %v"
	ErrorMin 	= "must be >= %s"
	ErrorMax 	= "must be <= %s"
)

type MinMaxValidator struct {
	min, max 	   string
	minVal, maxVal int
}

func NewMinMaxValidator(params map[string]string) Validator {
	validator := MinMaxValidator{}

	if minVal, ok := params["min"]; ok {
		validator.min = params["min"]
		if val, err := strconv.Atoi(minVal); err != nil {
			slog.Warn("NewMinMaxValidator invalid min limit provided", slog.String("min value", minVal))
		} else {
			validator.minVal = val
		}
	}

	if maxVal, ok := params["max"]; ok {
		validator.max = params["max"]
		if val, err := strconv.Atoi(maxVal); err != nil {
			slog.Warn("NewMinMaxValidator invalid max limit provided", slog.String("max value", maxVal))
		} else {
			validator.maxVal = val
		}
	}

	return validator
}

func (s MinMaxValidator) okMin(val int) bool {
	return val >= s.minVal
}

func (s MinMaxValidator) okMax(val int) bool {
	return val <= s.maxVal
}

func (s MinMaxValidator) Validate(value any) utils.Errors {
	var (
		val  int
		ok   bool
		errs = utils.Errors{}
	)

	if val, ok = value.(int); !ok {
		return append(errs, fmt.Errorf(ErrorNonInt, value))
	}

	if !s.okMin(val) {
		return append(errs, fmt.Errorf(ErrorMin, s.min))
	}

	if !s.okMax(val) {
		return append(errs, fmt.Errorf(ErrorMax, s.max))
	}
	
	if len(errs) > 0 {
		return errs
	}

	return nil
}