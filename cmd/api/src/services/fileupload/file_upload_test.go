package fileupload

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/buffer"
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
			badJSON = strings.NewReader("{\"redPill\": true, \"bluePill\": false}")
		)
		err := WriteAndValidateJSON(badJSON, &writer)
		assert.Nil(t, err)
	})
}
