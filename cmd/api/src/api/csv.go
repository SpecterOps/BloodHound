package api

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/utils"
)

var (
	ErrContentTypeCSV = errors.New("content type must be text/csv")
)

func ReadAPIV2CSVPayload(records *([][]string), response *http.Response) error {
	if !utils.HeaderMatches(response.Header, headers.ContentType.String(), mediatypes.TextCsv.String()) {
		return ErrContentTypeCSV
	}

	csvReader := csv.NewReader(response.Body)

	if content, err := csvReader.ReadAll(); err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	} else {
		*records = content
		return nil
	}

}
