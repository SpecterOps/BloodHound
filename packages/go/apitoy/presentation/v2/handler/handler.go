package handler

import (
	"mime"
	"net/http"
	"slices"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/packages/go/apitoy/app"
	"github.com/specterops/bloodhound/packages/go/apitoy/presentation/common"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/ingest"
)

// Handler stores dependencies of all handlers (currently just the BHEApp interface)
type Handler struct {
	bhApp app.BHApp
}

// NewHandler initializes a Handlers struct with a BHEApp
func NewHandler(bhApp app.BHApp) Handler {
	return Handler{
		bhApp: bhApp,
	}
}

const FileUploadJobIdPathParameterName = "file_upload_job_id"

func (s Handler) ProcessFileUpload(response http.ResponseWriter, request *http.Request) {
	var (
		requestId             = ctx.FromRequest(request).RequestID
		fileUploadJobIdString = mux.Vars(request)[FileUploadJobIdPathParameterName]
		contentType           = request.Header.Get(headers.ContentType.String())
	)

	if request.Body != nil {
		defer request.Body.Close()
	}

	if fileType := validateContentTypeForUpload(contentType); fileType == model.FileTypeInvalid {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Content type must be application/json or application/zip", request), response)
	} else if fileUploadJobID, err := strconv.Atoi(fileUploadJobIdString); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if err := s.bhApp.IngestFile(request.Context(), requestId, fileUploadJobID, fileType, request.Body); err != nil {
		common.WriteErrorResponse(request.Context(), request, response, err)
	} else {
		response.WriteHeader(http.StatusAccepted)
	}
}

func validateContentTypeForUpload(contentType string) model.FileType {
	if parsed, _, err := mime.ParseMediaType(contentType); err != nil {
		return model.FileTypeInvalid
	} else if parsed == mediatypes.ApplicationJson.String() {
		return model.FileTypeJson
	} else if slices.Contains(ingest.AllowedZipFileUploadTypes, parsed) {
		return model.FileTypeZip
	} else {
		return model.FileTypeInvalid
	}
}
