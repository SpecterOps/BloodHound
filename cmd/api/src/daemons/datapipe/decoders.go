package datapipe

import (
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/log"
	"io"
)

type ConversionFunc[T any] func(decoded T, converted *ConvertedData)

func decodeBasicData[T any](batch graph.Batch, reader io.ReadSeeker, conversionFunc ConversionFunc[T]) error {
	decoder, err := CreateIngestDecoder(reader)
	if err != nil {
		return err
	}

	var (
		count         = 0
		batchCount    = 0
		convertedData ConvertedData
	)

	for decoder.More() {
		var decodeTarget T
		if err := decoder.Decode(&decodeTarget); err != nil {
			log.Errorf("Error decoding %T object: %v", decodeTarget, err)
		} else {
			count++
			conversionFunc(decodeTarget, &convertedData)
		}

		if count == ingestCountThreshold {
			batchCount++
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

	convertedData := ConvertedGroupData{}
	var group ein.Group
	count := 0
	for decoder.More() {
		if err := decoder.Decode(&group); err != nil {
			log.Errorf("Error decoding group object: %v", err)
		} else {
			count++
			convertGroupData(group, &convertedData)
			if count == ingestCountThreshold {
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

	convertedData := ConvertedSessionData{}
	var session ein.Session
	count := 0
	for decoder.More() {
		if err := decoder.Decode(&session); err != nil {
			log.Errorf("Error decoding session object: %v", err)
		} else {
			count++
			convertSessionData(session, &convertedData)
			if count == ingestCountThreshold {
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

	convertedData := ConvertedAzureData{}
	var data AzureBase
	count := 0
	for decoder.More() {
		if err := decoder.Decode(&data); err != nil {
			log.Errorf("Error decoding azure object: %v", err)
		} else {
			convert := getKindConverter(data.Kind)
			convert(data.Data, &convertedData)
			count++
			if count == ingestCountThreshold {
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
