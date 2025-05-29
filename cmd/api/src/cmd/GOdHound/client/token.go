package client

import (
	"github.com/specterops/bloodhound/src/api"
	"net/http"
)

type TokenCredentialsHandler struct {
	TokenID  string
	TokenKey string
}

func (s *TokenCredentialsHandler) Handle(request *http.Request) error {
	return api.SignRequest(s.TokenID, s.TokenKey, request)
}
