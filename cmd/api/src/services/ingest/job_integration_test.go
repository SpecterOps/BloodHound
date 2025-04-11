package ingest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/specterops/bloodhound/src/model"
)

func TestIntegration_SaveIngestFile_JSON(t *testing.T) {
	// TODO: Make this
	body := bytes.NewBufferString(`{"meta":{"methods":46067,"type":"computers","count":0,"version":6},"data":[]}`)

	req := httptest.NewRequest(http.MethodPost, "/ingest", body)
	req.Header.Set("Content-Type", "application/json")

	// TODO: Make this take it a FS?
	location, fileType, err := SaveIngestFile("/tmp", req)
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
