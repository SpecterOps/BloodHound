package validation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/src/utils/validation"
)

func TestUrlValidator(t *testing.T) {
	t.Run("Valid URL", func(t *testing.T) {
		type testStruct struct {
			URL string `validate:"url"`
		}

		errs := validation.Validate(&testStruct{URL: "http://test.com"})
		require.Len(t, errs, 0)
	})

	t.Run("Invalid HTTP URL", func(t *testing.T) {
		type testStruct struct {
			URL string `validate:"url"`
		}
		errs := validation.Validate(&testStruct{URL: "bloodhound"})
		require.Len(t, errs, 1)
		assert.ErrorContains(t, errs[0], validation.ErrorUrlInvalid)
	})

	t.Run("Invalid HTTPS URL", func(t *testing.T) {
		type testStruct struct {
			URL string `validate:"url,httpsOnly=true"`
		}
		errs := validation.Validate(&testStruct{URL: "http://bloodhound.com"})
		require.Len(t, errs, 1)
		assert.ErrorContains(t, errs[0], validation.ErrorUrlHttpsInvalid)
	})
}
