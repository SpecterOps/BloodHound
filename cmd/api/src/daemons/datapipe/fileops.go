package datapipe

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func SeekToDataTag(decoder *json.Decoder) error {
	depth := 0
	dataTagFound := false
	for {
		if token, err := decoder.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				return ErrDataNotFound
			}

			return err
		} else {
			//Break here to allow for one more token read, which should take us to the "[" token, exactly where we need to be
			if dataTagFound {
				//Do some extra checks
				if typed, ok := token.(json.Delim); !ok {
					return ErrInvalidDataTag
				} else if typed != delimOpenSquareBracket {
					return ErrInvalidDataTag
				}
				//Break out of our loop if we're in a good spot
				return nil
			}
			switch typed := token.(type) {
			case json.Delim:
				switch typed {
				case delimCloseBracket, delimCloseSquareBracket:
					depth--
				case delimOpenBracket, delimOpenSquareBracket:
					depth++
				}
			case string:
				if !dataTagFound && depth == 1 && typed == "data" {
					dataTagFound = true
				}
			}
		}
	}
}

func CreateIngestDecoder(reader io.ReadSeeker) (*json.Decoder, error) {
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("error seeking to start of file: %w", err)
	} else {
		decoder := json.NewDecoder(reader)
		if err := SeekToDataTag(decoder); err != nil {
			return nil, fmt.Errorf("error seeking to data tag: %w", err)
		} else {
			return decoder, nil
		}
	}
}
