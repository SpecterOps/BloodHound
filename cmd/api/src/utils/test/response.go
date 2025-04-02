package test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/specterops/bloodhound/src/utils"
)

func ProcessResponse(t *testing.T, response *httptest.ResponseRecorder) (int, http.Header, []byte) {
	t.Helper()
	if response.Code != http.StatusOK && response.Code != http.StatusAccepted {
		responseBytes, err := utils.ReplaceFieldValueInJsonString(response.Body.String(), "timestamp", "0001-01-01T00:00:00Z")
		if err != nil {
			// not every error response contains a timestamp so print output and move along
			fmt.Printf("error replacing field value in json string: %v\n", err)
		}

		response.Body = bytes.NewBuffer([]byte(responseBytes))
	}

	if response.Body != nil {
		res, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatalf("error reading response body: %v", err)
		}

		return response.Code, response.Header(), res
	}

	return response.Code, response.Header(), nil
}
