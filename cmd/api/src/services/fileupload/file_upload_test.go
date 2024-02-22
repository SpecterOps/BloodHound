package fileupload

import (
	"github.com/specterops/bloodhound/src/model/ingest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/buffer"
	"io"
	"os"
	"strings"
	"testing"
)

func TestWriteAndValidateJSON(t *testing.T) {
	t.Run("trigger invalid json on bad json", func(t *testing.T) {
		var (
			writer  = buffer.Buffer{}
			badJSON = strings.NewReader("{[]}")
		)
		err := WriteAndValidateJSON(badJSON, &writer)
		assert.ErrorIs(t, err, ErrInvalidJSON)
	})

	t.Run("succeed on good json", func(t *testing.T) {
		var (
			writer  = buffer.Buffer{}
			badJSON = strings.NewReader(`{"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}, "data": []}`)
		)
		err := WriteAndValidateJSON(badJSON, &writer)
		assert.Nil(t, err)
	})
}

func TestWriteAndValidateZip(t *testing.T) {
	t.Run("valid zip file is ok", func(t *testing.T) {
		var (
			writer = buffer.Buffer{}
		)

		file, err := os.Open("../../test/fixtures/fixtures/goodzip.zip")
		assert.Nil(t, err)

		err = WriteAndValidateZip(io.Reader(file), &writer)
		assert.Nil(t, err)
	})

	t.Run("invalid bytes causes error", func(t *testing.T) {
		var (
			writer = buffer.Buffer{}
			badZip = strings.NewReader("123123")
		)

		err := WriteAndValidateZip(badZip, &writer)
		assert.Equal(t, err, ingest.ErrInvalidZipFile)
	})
}
