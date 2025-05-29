package client

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	"net/http"
)

type SecretCredentialsHandler struct {
	Username string
	Secret   string
	Client   Client
	jwt      string
	session  auth.SessionData
}

func (s *SecretCredentialsHandler) SetSessionToken(sessionToken string) error {
	var jwtParser jwt.Parser

	if _, _, err := jwtParser.ParseUnverified(sessionToken, &s.session); err != nil {
		return fmt.Errorf("failed pasring JWT session token: %w", err)
	} else {
		s.jwt = sessionToken
	}

	return nil
}

func (s *SecretCredentialsHandler) login() error {
	if resp, err := s.Client.LoginSecret(s.Username, s.Secret); err != nil {
		return err
	} else {
		return s.SetSessionToken(resp.SessionToken)
	}
}

func (s *SecretCredentialsHandler) Handle(request *http.Request) error {
	if s.jwt == "" || s.session.Valid() != nil {
		if err := s.login(); err != nil {
			return err
		}
	}

	request.Header.Set(headers.Authorization.String(), fmt.Sprintf("%s %s", api.AuthorizationSchemeBearer, s.jwt))
	return nil
}

