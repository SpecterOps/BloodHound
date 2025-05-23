// Copyright 2024 Specter Ops, Inc.
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

package graphify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/specterops/bloodhound/dawgs/util"
	"github.com/specterops/bloodhound/ein"
)

type ConversionFuncWithTime[T any] func(decoded T, converted *ConvertedData, ingestTime time.Time)

// ConversionFunc is a function that transforms a decoded JSON object (of type T)
// into its corresponding internal ingest representation, appending it to the provided ConvertedData.
//
// T represents a specific ingest type (e.g., User, Computer, Group, etc.).
type ConversionFunc[T any] func(decoded T, converted *ConvertedData) error

func decodeBasicData[T any](batch *TimestampedBatch, decoder *json.Decoder, conversionFunc ConversionFuncWithTime[T]) error {
	var (
		count         = 0
		convertedData ConvertedData
		errs          = util.NewErrorCollector()
	)

	for decoder.More() {
		// This variable needs to be initialized here, otherwise the marshaller will cache the map in the struct
		var decodeTarget T
		if err := decoder.Decode(&decodeTarget); err != nil {
			slog.Error(fmt.Sprintf("Error decoding %T object: %v", decodeTarget, err))
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		} else {
			count++
			conversionFunc(decodeTarget, &convertedData, batch.IngestTime)
		}

		if count == IngestCountThreshold {
			if err := IngestBasicData(batch, convertedData); err != nil {
				errs.Add(err)
			}
			convertedData.Clear()
			count = 0

		}
	}

	if count > 0 {
		if err := IngestBasicData(batch, convertedData); err != nil {
			errs.Add(err)
		}
	}

	return errs.Combined()
}

func decodeGenericData[T any](batch *TimestampedBatch, decoder *json.Decoder, conversionFunc ConversionFunc[T]) error {
	var (
		count         = 0
		convertedData ConvertedData
		errs          = util.NewErrorCollector()
	)

	for decoder.More() {
		// This variable needs to be initialized here, otherwise the marshaller will cache the map in the struct
		var decodeTarget T
		if err := decoder.Decode(&decodeTarget); err != nil {
			slog.Error(fmt.Sprintf("Error decoding %T object: %v", decodeTarget, err))
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		} else {
			count++
			if err := conversionFunc(decodeTarget, &convertedData); err != nil {
				errs.Add(err)
			}
		}

		if count == IngestCountThreshold {
			if err := IngestGenericData(batch, convertedData); err != nil {
				errs.Add(err)
			}
			convertedData.Clear()
			count = 0

		}
	}

	if count > 0 {
		if err := IngestGenericData(batch, convertedData); err != nil {
			errs.Add(err)
		}
	}

	return errs.Combined()
}

func decodeGroupData(batch *TimestampedBatch, decoder *json.Decoder) error {

	var (
		convertedData = ConvertedGroupData{}
		count         = 0
		errs          = util.NewErrorCollector()
	)

	for decoder.More() {
		var group ein.Group
		if err := decoder.Decode(&group); err != nil {
			slog.Error(fmt.Sprintf("Error decoding group object: %v", err))
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		} else {
			count++
			convertGroupData(group, &convertedData, batch.IngestTime)
			if count == IngestCountThreshold {
				if err = IngestGroupData(batch, convertedData); err != nil {
					errs.Add(err)
				}

				convertedData.Clear()
				count = 0
			}
		}
	}

	if count > 0 {
		if err := IngestGroupData(batch, convertedData); err != nil {
			errs.Add(err)
		}
	}

	return errs.Combined()
}

func decodeSessionData(batch *TimestampedBatch, decoder *json.Decoder) error {
	var (
		convertedData = ConvertedSessionData{}
		count         = 0
		errs          = util.NewErrorCollector()
	)

	for decoder.More() {
		var session ein.Session
		if err := decoder.Decode(&session); err != nil {
			slog.Error(fmt.Sprintf("Error decoding session object: %v", err))
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		} else {
			count++
			convertSessionData(session, &convertedData)
			if count == IngestCountThreshold {
				if err = IngestSessions(batch, convertedData.SessionProps); err != nil {
					errs.Add(err)
				}
				convertedData.Clear()
				count = 0
			}
		}
	}

	if count > 0 {
		if err := IngestSessions(batch, convertedData.SessionProps); err != nil {
			errs.Add(err)
		}
	}

	return errs.Combined()
}

func decodeAzureData(batch *TimestampedBatch, decoder *json.Decoder) error {
	var (
		convertedData = ConvertedAzureData{}
		count         = 0
		errs          = util.NewErrorCollector()
	)

	for decoder.More() {
		var data AzureBase
		if err := decoder.Decode(&data); err != nil {
			slog.Error(fmt.Sprintf("Error decoding azure object: %v", err))
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		} else {
			convert := getKindConverter(data.Kind)
			convert(data.Data, &convertedData, batch.IngestTime)
			count++
			if count == IngestCountThreshold {
				if err = IngestAzureData(batch, convertedData); err != nil {
					errs.Add(err)
				}
				convertedData.Clear()
				count = 0
			}
		}
	}

	if count > 0 {
		if err := IngestAzureData(batch, convertedData); err != nil {
			errs.Add(err)
		}
	}

	return errs.Combined()
}
