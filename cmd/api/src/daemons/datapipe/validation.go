package datapipe

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func ValidateMetaTag(reader io.ReadSeeker) (Metadata, error) {
	_, err := reader.Seek(0, io.SeekStart)
	if err != nil {
		return Metadata{}, fmt.Errorf("error seeking to start of file: %w", err)
	}
	depth := 0
	decoder := json.NewDecoder(reader)
	dataTagFound := false
	metaTagFound := false
	var meta Metadata
	for {
		if dataTagFound && metaTagFound {
			return meta, nil
		}
		if token, err := decoder.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				if !metaTagFound && !dataTagFound {
					return Metadata{}, ErrNoTagFound
				} else if !metaTagFound {
					return Metadata{}, ErrDataNotFound
				} else {
					return Metadata{}, ErrMetaNotFound
				}
			}
			return Metadata{}, err
		} else {
			switch typed := token.(type) {
			case json.Delim:
				switch typed {
				case delimCloseBracket, delimCloseSquareBracket:
					depth--
				case delimOpenBracket, delimOpenSquareBracket:
					depth++
				}
			case string:
				if !metaTagFound && depth == 1 && typed == "meta" {
					if err := decoder.Decode(&meta); err != nil {
						return Metadata{}, err
					} else if meta.IsValid() {
						metaTagFound = true
					}
				}

				if !dataTagFound && depth == 1 && typed == "data" {
					dataTagFound = true
				}
			}
		}
	}
}
