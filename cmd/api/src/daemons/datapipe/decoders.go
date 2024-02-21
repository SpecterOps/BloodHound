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

package datapipe

import (
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/log"
	"io"
)

/*
ConversionFunc is responsible for turning an individual json object into the equivalent ingest object and storing the data into ConvertedData.

T is any of the ingest types
*/
type ConversionFunc[T any] func(decoded T, converted *ConvertedData)

func decodeBasicData[T any](batch graph.Batch, reader io.ReadSeeker, conversionFunc ConversionFunc[T]) error {
	decoder, err := CreateIngestDecoder(reader)
	if err != nil {
		return err
	}

	var (
		count         = 0
		convertedData ConvertedData
	)

	for decoder.More() {
		//This variable needs to be initialized here, otherwise the marshaller will cache the map in the struct
		var decodeTarget T
		if err := decoder.Decode(&decodeTarget); err != nil {
			log.Errorf("Error decoding %T object: %v", decodeTarget, err)
		} else {
			count++
			conversionFunc(decodeTarget, &convertedData)
		}

		if count == IngestCountThreshold {
			IngestBasicData(batch, convertedData)
			convertedData.Clear()
			count = 0
		}
	}

	if count > 0 {
		IngestBasicData(batch, convertedData)
	}

	return nil
}

func decodeGroupData(batch graph.Batch, reader io.ReadSeeker) error {
	decoder, err := CreateIngestDecoder(reader)
	if err != nil {
		return err
	}

	var (
		convertedData = ConvertedGroupData{}
		group         ein.Group
		count         = 0
	)

	for decoder.More() {
		if err := decoder.Decode(&group); err != nil {
			log.Errorf("Error decoding group object: %v", err)
		} else {
			count++
			convertGroupData(group, &convertedData)
			if count == IngestCountThreshold {
				IngestGroupData(batch, convertedData)
				convertedData.Clear()
				count = 0
			}
		}
	}

	if count > 0 {
		IngestGroupData(batch, convertedData)
	}

	return nil
}

func decodeSessionData(batch graph.Batch, reader io.ReadSeeker) error {
	decoder, err := CreateIngestDecoder(reader)
	if err != nil {
		return err
	}

	var (
		convertedData = ConvertedSessionData{}
		session       ein.Session
		count         = 0
	)
	for decoder.More() {
		if err := decoder.Decode(&session); err != nil {
			log.Errorf("Error decoding session object: %v", err)
		} else {
			count++
			convertSessionData(session, &convertedData)
			if count == IngestCountThreshold {
				IngestSessions(batch, convertedData.SessionProps)
				convertedData.Clear()
				count = 0
			}
		}
	}

	if count > 0 {
		IngestSessions(batch, convertedData.SessionProps)
	}

	return nil
}

func decodeAzureData(batch graph.Batch, reader io.ReadSeeker) error {
	decoder, err := CreateIngestDecoder(reader)
	if err != nil {
		return err
	}

	var (
		convertedData = ConvertedAzureData{}
		data          AzureBase
		count         = 0
	)

	for decoder.More() {
		if err := decoder.Decode(&data); err != nil {
			log.Errorf("Error decoding azure object: %v", err)
		} else {
			convert := getKindConverter(data.Kind)
			convert(data.Data, &convertedData)
			count++
			if count == IngestCountThreshold {
				IngestAzureData(batch, convertedData)
				convertedData.Clear()
				count = 0
			}
		}
	}

	if count > 0 {
		IngestAzureData(batch, convertedData)
	}

	return nil
}
