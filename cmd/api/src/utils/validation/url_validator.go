package validation

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/specterops/bloodhound/src/utils"
)

const (
	ErrorUrlHttpsInvalid = "invalid https url format"
	ErrorUrlInvalid      = "invalid url format"
)

// UrlValidator implements the Validator interface which allows the usage of the `url` struct tag
// to ensure a string is in a valid url format by calling `Validate`
type UrlValidator struct {
	forceHttps bool
}

// NewUrlValidator returns a new Validator
func NewUrlValidator(params map[string]string) Validator {
	validator := UrlValidator{}

	if val, ok := params["https"]; ok {
		validator.forceHttps, _ = strconv.ParseBool(val)
	}

	return validator
}

// Validate validates that the associated struct fields are in the proper formatting
func (s UrlValidator) Validate(value any) utils.Errors {
	var (
		errs         = utils.Errors{}
		inputUrl, ok = value.(string)
	)

	if !ok {
		errs = append(errs, fmt.Errorf("expected a string value, got %s", reflect.TypeOf(value)))
	}

	if s.forceHttps {
		if !strings.HasPrefix(inputUrl, "https://") {
			errs = append(errs, errors.New(ErrorUrlHttpsInvalid))
		}
	}

	if err := validUrl(inputUrl); err != nil {
		errs = append(errs, errors.New(ErrorUrlInvalid))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func validUrl(inputUrl string) error {
	if _, err := url.ParseRequestURI(inputUrl); err != nil {
		return err
	}

	return nil
}
