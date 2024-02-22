package datapipe_test

import (
	"encoding/json"
	"github.com/specterops/bloodhound/src/model/ingest"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/src/daemons/datapipe"
	"github.com/stretchr/testify/assert"
)

type dataTagAssertion struct {
	rawString string
	err       error
}

func TestSeekToDataTag(t *testing.T) {
	assertions := []dataTagAssertion{
		{
			rawString: "{\"data\": []}",
			err:       nil,
		},
		{
			rawString: "{\"data\": {}}",
			err:       ingest.ErrInvalidDataTag,
		},
		{
			rawString: "{\"data\": ]}",
			err:       ingest.ErrJSONDecoderInternal,
		},
		{
			rawString: "",
			err:       ingest.ErrDataTagNotFound,
		},
		{
			rawString: "{[]}",
			err:       ingest.ErrJSONDecoderInternal,
		},
		{
			rawString: "{\"data\": \"oops\"}",
			err:       ingest.ErrInvalidDataTag,
		},
		{
			rawString: "{\"nothing\": [}",
			err:       ingest.ErrJSONDecoderInternal,
		},
		{
			rawString: `{"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}, "data": []}`,
			err:       nil,
		},
		{
			rawString: `{"test": {"data": {}}, "meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}, "data": []}`,
			err:       nil,
		},
	}

	for _, assertion := range assertions {
		r := strings.NewReader(assertion.rawString)
		j := json.NewDecoder(r)

		err := datapipe.SeekToDataTag(j)
		assert.ErrorIs(t, err, assertion.err)
	}
}
