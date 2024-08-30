package validation_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/specterops/bloodhound/src/utils"
	"github.com/specterops/bloodhound/src/utils/validation"
)

func TestDurationValidator(t *testing.T) {

	type Epic struct {
		Duration time.Duration `validate:"duration,min=P1D,max=P14D"`
	}

	fortnight := time.Hour * 24 * 14

	minErr := fmt.Errorf("Duration: "+validation.ErrorDurationMin, "P1D")
	maxErr := fmt.Errorf("Duration: "+validation.ErrorDurationMax, "P14D")

	var cases = []struct {
		Input  Epic
		Errors utils.Errors
	}{
		{Epic{time.Hour}, utils.Errors{minErr}},
		{Epic{fortnight}, nil},
		{Epic{fortnight + time.Hour}, utils.Errors{maxErr}},
	}

	for _, tc := range cases {
		if errs := validation.Validate(tc.Input); errs != nil {
			if tc.Errors == nil {
				t.Errorf("For input: %v, expected errs to be nil: errs = %v\n", tc.Input, errs)
			} else if errs.Error() != tc.Errors.Error() {
				t.Errorf("For input: %v, got %v, want %v\n", tc.Input, errs, tc.Errors)
			}
		} else if tc.Errors != nil {
			t.Errorf("For input: %v, expected errs to not be nil\n", tc.Input)
		}
	}
}
