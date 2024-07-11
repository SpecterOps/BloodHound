package app

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/ingest"
)

var ErrNotFound = errors.New("not found")
var ErrFileValidation = errors.New("file validation failure")
var ErrGenericDatabaseFailure = errors.New("database failure")
var ErrInvalidJSON = errors.New("file is not valid json")
var ErrGeneralApplicationFailure = errors.New("general application failure")

type BHApp struct {
	// adapter adapter.PostgresAdapter
	db  database.Database
	cfg config.Configuration
}

// NewBHEApp ingests BHEAppServices into a new BHEApp service container struct
func NewBHApp(db database.Database, cfg config.Configuration) BHApp {
	return BHApp{
		db:  db,
		cfg: cfg,
	}
}

func (s BHApp) IngestFile(ctx context.Context, requestID string, jobID int, fileType model.FileType, content io.ReadCloser) error {
	var validationStrategy FileValidator

	switch fileType {
	case model.FileTypeJson:
		validationStrategy = WriteAndValidateJSON
	case model.FileTypeZip:
		validationStrategy = WriteAndValidateZip
	default:
		return ErrFileValidation
	}

	if fileUploadJob, err := getFileUploadJobByID(ctx, s.db, jobID); err != nil {
		return err
	} else if tempFile, err := saveIngestFile(s.cfg.TempDirectory(), content, validationStrategy); err != nil {
		return err
	} else if _, err := createIngestTask(ctx, s.db, tempFile, fileType, requestID, jobID); err != nil {
		return err
	} else if err := touchFileUploadJobLastIngest(ctx, s.db, fileUploadJob); err != nil {
		return err
	} else {
		return nil
	}
}

func getFileUploadJobByID(ctx context.Context, db database.Database, jobID int) (model.FileUploadJob, error) {
	if job, err := db.GetFileUploadJob(ctx, int64(jobID)); errors.Is(err, database.ErrNotFound) {
		return job, fmt.Errorf("get file upload job by id: %w: %v", ErrNotFound, err)
	} else if err != nil {
		return job, fmt.Errorf("get file upload job by id: %w: %v", ErrGenericDatabaseFailure, err)
	} else {
		return job, nil
	}
}

func createIngestTask(ctx context.Context, db database.Database, filename string, fileType model.FileType, requestID string, jobID int) (model.IngestTask, error) {
	newIngestTask := model.IngestTask{
		FileName:    filename,
		RequestGUID: requestID,
		TaskID:      null.Int64From(int64(jobID)),
		FileType:    fileType,
	}

	if task, err := db.CreateIngestTask(ctx, newIngestTask); err != nil {
		return task, fmt.Errorf("create ingest task: %w: %v", ErrGenericDatabaseFailure, err)
	} else {
		return task, nil
	}
}

func touchFileUploadJobLastIngest(ctx context.Context, db database.Database, fileUploadJob model.FileUploadJob) error {
	fileUploadJob.LastIngest = time.Now().UTC()
	if err := db.UpdateFileUploadJob(ctx, fileUploadJob); err != nil {
		return fmt.Errorf("touch last ingest: %w: %v", ErrGenericDatabaseFailure, err)
	} else {
		return nil
	}
}

func saveIngestFile(tempDir string, body io.ReadCloser, validationStrategy FileValidator) (string, error) {
	tempFile, err := os.CreateTemp(tempDir, "bh")
	if err != nil {
		return "", fmt.Errorf("creating ingest file: %w: %v", ErrGeneralApplicationFailure, err)
	}

	if err := validationStrategy(body, tempFile); err != nil {
		if err := tempFile.Close(); err != nil {
			log.Errorf("Error closing temp file %s with failed validation: %v", tempFile.Name(), err)
		} else if err := os.Remove(tempFile.Name()); err != nil {
			log.Errorf("Error deleting temp file %s: %v", tempFile.Name(), err)
		}
		return tempFile.Name(), fmt.Errorf("saving ingest file: %w: %v", ErrFileValidation, err)
	} else {
		if err := tempFile.Close(); err != nil {
			log.Errorf("Error closing temp file with successful validation %s: %v", tempFile.Name(), err)
		}
		return tempFile.Name(), nil
	}
}

type FileValidator func(src io.Reader, dst io.Writer) error

var ZipMagicBytes = []byte{0x50, 0x4b, 0x03, 0x04}

// ValidateMetaTag ensures that the correct tags are present in a json file for data ingest.
// If readToEnd is set to true, the stream will read to the end of the file (needed for TeeReader)
func ValidateMetaTag(reader io.Reader, readToEnd bool) (ingest.Metadata, error) {
	var (
		depth            = 0
		decoder          = json.NewDecoder(reader)
		dataTagFound     = false
		dataTagValidated = false
		metaTagFound     = false
		meta             ingest.Metadata
	)

	for {
		if token, err := decoder.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				if !metaTagFound && !dataTagFound {
					return ingest.Metadata{}, ingest.ErrNoTagFound
				} else if !dataTagFound {
					return ingest.Metadata{}, ingest.ErrDataTagNotFound
				} else {
					return ingest.Metadata{}, ingest.ErrMetaTagNotFound
				}
			} else {
				return ingest.Metadata{}, ErrInvalidJSON
			}
		} else {
			//Validate that our data tag is actually opening correctly
			if dataTagFound && !dataTagValidated {
				if typed, ok := token.(json.Delim); ok && typed == ingest.DelimOpenSquareBracket {
					dataTagValidated = true
				} else {
					dataTagFound = false
				}
			}
			switch typed := token.(type) {
			case json.Delim:
				switch typed {
				case ingest.DelimCloseBracket, ingest.DelimCloseSquareBracket:
					depth--
				case ingest.DelimOpenBracket, ingest.DelimOpenSquareBracket:
					depth++
				}
			case string:
				if !metaTagFound && depth == 1 && typed == "meta" {
					if err := decoder.Decode(&meta); err != nil {
						log.Warnf("Found invalid metatag, skipping")
					} else if meta.Type.IsValid() {
						metaTagFound = true
					}
				}

				if !dataTagFound && depth == 1 && typed == "data" {
					dataTagFound = true
				}
			}
		}

		if dataTagValidated && metaTagFound {
			break
		}
	}

	if readToEnd {
		if _, err := io.Copy(io.Discard, reader); err != nil {
			return ingest.Metadata{}, err
		}
	}

	return meta, nil
}

func ValidateZipFile(reader io.Reader) error {
	bytes := make([]byte, 4)
	if readBytes, err := reader.Read(bytes); err != nil {
		return err
	} else if readBytes < 4 {
		return ingest.ErrInvalidZipFile
	} else {
		for i := 0; i < 4; i++ {
			if bytes[i] != ZipMagicBytes[i] {
				return ingest.ErrInvalidZipFile
			}
		}

		_, err := io.Copy(io.Discard, reader)

		return err
	}
}

func WriteAndValidateZip(src io.Reader, dst io.Writer) error {
	tr := io.TeeReader(src, dst)
	return ValidateZipFile(tr)
}

const (
	UTF8BOM1 = 0xef
	UTF8BOM2 = 0xbb
	UTF8BMO3 = 0xbf
)

func WriteAndValidateJSON(src io.Reader, dst io.Writer) error {
	tr := io.TeeReader(src, dst)
	bufReader := bufio.NewReader(tr)
	if b, err := bufReader.Peek(3); err != nil {
		return err
	} else {
		if b[0] == UTF8BOM1 && b[1] == UTF8BOM2 && b[2] == UTF8BMO3 {
			if _, err := bufReader.Discard(3); err != nil {
				return err
			}
		}
	}
	_, err := ValidateMetaTag(bufReader, true)
	return err
}
