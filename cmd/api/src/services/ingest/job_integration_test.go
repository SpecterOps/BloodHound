package ingest

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/specterops/bloodhound/src/model"
)

func TestIntegration_SaveIngestFile_JSON(t *testing.T) {
	// TODO: Make this
	body := io.NopCloser(bytes.NewBufferString(`{"meta":{"methods":46067,"type":"computers","count":0,"vers ion":6},"data":[]}`))

	// TODO: Make this take it a FS?
	location, fileType, err := SaveIngestFile("/tmp", "application/json", body)
	if err != nil {
		t.Fatalf("SaveIngestFile failed: %v", err)
	}

	if fileType != model.FileTypeJson {
		t.Errorf("expected fileType %v, got %v", model.FileTypeJson, fileType)
	}

	info, statErr := os.Stat(location)
	if statErr != nil {
		t.Fatalf("output file not found: %v", statErr)
	}

	if info.Size() == 0 {
		t.Error("file saved is empty")
	}

	// TODO: Make this unnecessary
	os.Remove(location)
}
