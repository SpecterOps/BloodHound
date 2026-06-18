package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
)

//go:generate go tool mockery

type Service interface {
	GetDatapipeStatus(context.Context) (services.DatapipeStatus, error)
}

type Handlers struct {
	service Service
}

func NewHandlers(service Service) *Handlers {
	return &Handlers{
		service: service,
	}
}

func (s *Handlers) GetDatapipeStatus(response http.ResponseWriter, request *http.Request) {
	var ctx = request.Context()

	status, err := s.service.GetDatapipeStatus(ctx)
	if err != nil {
		handleServiceError(request, response, err)
		return
	}

	responses.WriteBasic(ctx, BuildDatapipeStatusView(status), http.StatusOK, response)
}

func handleServiceError(request *http.Request, response http.ResponseWriter, err error) {
	var ctx = request.Context()

	if errors.Is(err, services.ErrNotFound) {
		responses.WriteError(ctx, http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, response)
	} else if errors.Is(err, context.DeadlineExceeded) {
		responses.WriteError(ctx, http.StatusInternalServerError, api.ErrorResponseRequestTimeout, response)
	} else {
		slog.ErrorContext(ctx, "Unexpected database error", attr.Error(err))
		responses.WriteError(ctx, http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, response)
	}
}
