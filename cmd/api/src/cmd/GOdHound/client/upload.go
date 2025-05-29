package client

import (
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"io"
	"net/http"
	"strconv"

	"github.com/specterops/bloodhound/src/api"
)

func (s Client) Version() (v2.VersionResponse, error) {
	if resp, err := s.AuthedRequest(http.MethodGet, "/api/version", nil, nil); err != nil {
		return v2.VersionResponse{}, err
	} else {
		defer resp.Body.Close()

		var uploadJob v2.VersionResponse

		if api.IsErrorResponse(resp) {
			return uploadJob, ReadAPIError(resp)
		}

		return uploadJob, api.ReadAPIV2ResponsePayload(&uploadJob, resp)
	}
}

func (s Client) StartFileUploadJob() (FileUploadJob, error) {
	if resp, err := s.AuthedRequest(http.MethodPost, "/api/v2/file-upload/start", nil, nil); err != nil {
		return FileUploadJob{}, err
	} else {
		defer resp.Body.Close()

		var uploadJob FileUploadJob

		if api.IsErrorResponse(resp) {
			return uploadJob, ReadAPIError(resp)
		}

		return uploadJob, api.ReadAPIV2ResponsePayload(&uploadJob, resp)
	}
}

func (s Client) SendFileUploadPart(job FileUploadJob, reader io.Reader) error {
	var (
		jobPath = "/api/v2/file-upload/" + strconv.FormatInt(job.ID, 10)
		headers = http.Header{
			"Content-Type": []string{"application/json"},
		}
	)

	if resp, err := s.AuthedRequest(http.MethodPost, jobPath, nil, reader, headers); err != nil {
		return err
	} else {
		defer resp.Body.Close()

		if api.IsErrorResponse(resp) {
			return ReadAPIError(resp)
		}

		return nil
	}
}

func (s Client) EndFileUploadJob(job FileUploadJob) error {
	var (
		jobPath   = "/api/v2/file-upload/" + strconv.FormatInt(job.ID, 10) + "/end"
		resp, err = s.AuthedRequest(http.MethodPost, jobPath, nil, nil)
	)

	defer resp.Body.Close()
	return err
}
