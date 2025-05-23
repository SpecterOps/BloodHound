package validation_test

import (
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/src/utils"
	"github.com/specterops/bloodhound/src/utils/validation"
)

func TestRRuleValidator(t *testing.T) {

	type Schedule struct {
		RRule string `validate:"rrule"`
	}

	var cases = []struct {
		Input  Schedule
		Errors utils.Errors
	}{
		{Schedule{"abc123"}, utils.Errors{fmt.Errorf("RRule: invalid rrule specified: wrong format")}},
		{Schedule{"FREQ=DAILY;INTERVAL=1;COUNT=3"}, utils.Errors{fmt.Errorf("RRule: invalid rrule specified: count not supported")}},
		{Schedule{"FREQ=DAILY;INTERVAL=1;UNTIL=20240930T000000Z"}, utils.Errors{fmt.Errorf("RRule: invalid rrule specified: until not supported")}},
		{Schedule{"RRULE:FREQ=DAILY;INTERVAL=1"}, utils.Errors{fmt.Errorf("RRule: invalid rrule specified: dtstart is required")}},
		{Schedule{"DTSTART:20140909T100000Z\nRRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=TU"}, nil},
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
