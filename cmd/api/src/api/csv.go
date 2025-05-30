// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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
