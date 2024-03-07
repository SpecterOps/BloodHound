package openapi

import (
	_ "embed"
	"net/http"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
)

//go:embed doc/openapi.json
var openApiJson []byte

func HttpHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	w.WriteHeader(http.StatusOK)
	w.Write(openApiJson)
}
