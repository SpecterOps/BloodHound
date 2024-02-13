package datapipe_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/src/daemons/datapipe"
	"github.com/stretchr/testify/assert"
)

type dataTagAssertion struct {
	rawString string
	err       error
}

// TODO: probably some more complex cases worth considering, ideally something that we could expect would be close to a real JSON object, but variations that should fail too.
func TestSeekToDataTag(t *testing.T) {
	assertions := []dataTagAssertion{
		{
			rawString: "{\"data\": []}",
			err:       nil,
		},
		{
			rawString: "{\"data\": {}}",
			err:       datapipe.ErrInvalidDataTag,
		},
		{
			rawString: "{\"data\": ]}",
			err:       datapipe.ErrJSONDecoderInternal,
		},
		{
			rawString: "",
			err:       datapipe.ErrDataTagNotFound,
		},
		{
			rawString: "{[]}",
			err:       datapipe.ErrJSONDecoderInternal,
		},
		{
			rawString: "{\"data\": \"oops\"}",
			err:       datapipe.ErrInvalidDataTag,
		},
		{
			rawString: "{\"nothing\": [}",
			err:       datapipe.ErrJSONDecoderInternal,
		},
	}

	for _, assertion := range assertions {
		r := strings.NewReader(assertion.rawString)
		j := json.NewDecoder(r)

		err := datapipe.SeekToDataTag(j)
		assert.ErrorIs(t, err, assertion.err)
	}
}
