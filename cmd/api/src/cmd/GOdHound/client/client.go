package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/specterops/bloodhound/src/api/v2/apiclient"
	"github.com/specterops/bloodhound/src/auth"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api"
)

type CredentialsHandler interface {
	Handle(request *http.Request) error
}

type Client struct {
	Credentials CredentialsHandler
	ServiceURL  url.URL
	httpClient  *http.Client
}

func NewClient(rawServiceURL string, credentials apiclient.CredentialsHandler) (Client, error) {
	if serviceURL, err := url.Parse(rawServiceURL); err != nil {
		return Client{}, err
	} else {
		newHttpClient := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2: true,
				MaxIdleConns:      100,
			},
		}

		newClient := Client{
			Credentials: credentials,
			httpClient:  newHttpClient,
			ServiceURL:  *serviceURL,
		}

		switch typedCredentials := credentials.(type) {
		case *SecretCredentialsHandler:
			typedCredentials.Client = newClient
		}

		return newClient, nil
	}
}

func (s Client) ZipRequest(method, path string, params url.Values, body []byte) (*http.Response, error) {
	endpoint := api.URLJoinPath(s.ServiceURL, path)
	endpoint.RawQuery = params.Encode()

	if request, err := http.NewRequest(method, endpoint.String(), nil); err != nil {
		return nil, err
	} else {
		// query the Request and hand the response back to the user
		const (
			sleepInterval = time.Second * 5
			maxSleep      = sleepInterval * 5
		)

		started := time.Now()

		for {
			// Serialize the Request body - we expect a JSON serializable object here
			// This must be done on every retry, otherwise the buffer will be empty because it had been read
			if body != nil {
				request.Header.Set(headers.ContentType.String(), mediatypes.ApplicationZip.String())
				request.Body = io.NopCloser(bytes.NewReader(body))
			}

			// Set our credentials either via signage or bearer token session
			// Credentials also have to be set with every attempt due to request signing
			if s.Credentials != nil {
				if err := s.Credentials.Handle(request); err != nil {
					return nil, err
				}
			}

			if response, err := s.httpClient.Do(request); err != nil {
				if time.Since(started) >= maxSleep {
					return nil, fmt.Errorf("waited %f seconds while retrying - Request failure cause: %w", maxSleep.Seconds(), err)
				}

				slog.Error(fmt.Sprintf("Request to %s failed with error: %v. Attempting a retry.", endpoint.String(), err))
				time.Sleep(sleepInterval)
			} else {
				return response, nil
			}
		}
	}
}

func (s Client) Request(method, path string, params url.Values, body any, header ...http.Header) (*http.Response, error) {
	endpoint := api.URLJoinPath(s.ServiceURL, path)
	endpoint.RawQuery = params.Encode()

	if request, err := http.NewRequest(method, endpoint.String(), nil); err != nil {
		return nil, err
	} else {
		// Set the header
		if len(header) > 0 {
			request.Header = header[0]
		}

		// Serialize the Request body - we expect a JSON serializable object here
		// This must be done on every retry, otherwise the buffer will be empty because it had been read
		if body != nil {
			buffer := &bytes.Buffer{}

			if err := json.NewEncoder(buffer).Encode(body); err != nil {
				return nil, err
			}

			request.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
			request.Body = io.NopCloser(buffer)
		}

		return s.httpClient.Do(request)
	}
}

func (s Client) AuthedRequest(method, path string, params url.Values, body any, header ...http.Header) (*http.Response, error) {
	endpoint := api.URLJoinPath(s.ServiceURL, path)
	endpoint.RawQuery = params.Encode()

	if request, err := http.NewRequest(method, endpoint.String(), nil); err != nil {
		return nil, err
	} else {
		// Set the header
		if len(header) > 0 {
			request.Header = header[0]
		}

		// Serialize the Request body - we expect a JSON serializable object here
		// This must be done on every retry, otherwise the buffer will be empty because it had been read
		if body != nil {
			switch typedBody := body.(type) {
			case io.Reader:
				request.Body = io.NopCloser(typedBody)

			case io.ReadCloser:
				request.Body = typedBody

			default:
				buffer := &bytes.Buffer{}

				if err := json.NewEncoder(buffer).Encode(body); err != nil {
					return nil, err
				}

				request.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
				request.Body = io.NopCloser(buffer)
			}
		}

		// Set our credentials either via signage or bearer token session
		// Credentials also have to be set with every attempt due to request signing
		if s.Credentials != nil {
			if err := s.Credentials.Handle(request); err != nil {
				return nil, err
			}
		}

		return s.httpClient.Do(request)
	}
}

func (s Client) NewRequest(method string, path string, params url.Values, body io.ReadCloser, header ...http.Header) (*http.Request, error) {
	endpoint := api.URLJoinPath(s.ServiceURL, path)
	endpoint.RawQuery = params.Encode()

	request, err := http.NewRequest(method, endpoint.String(), body)

	if len(header) > 0 {
		request.Header = header[0]
	}

	return request, err
}

func (s Client) Raw(request *http.Request) (*http.Response, error) {
	// Set our credentials either via signage or bearer token session
	if s.Credentials != nil {
		if err := s.Credentials.Handle(request); err != nil {
			return nil, err
		}
	}

	// query the Request and hand the response back to the user
	return s.httpClient.Do(request)
}

func (s Client) LoginSecret(username, secret string) (api.LoginResponse, error) {
	var (
		loginResponse api.LoginResponse
		loginRequest  = api.LoginRequest{
			LoginMethod: auth.ProviderTypeSecret,
			Username:    username,
			Secret:      secret,
		}
	)

	if response, err := s.Request(http.MethodPost, "api/v2/login", nil, loginRequest); err != nil {
		return loginResponse, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return loginResponse, ReadAPIError(response)
		}

		return loginResponse, api.ReadAPIV2ResponsePayload(&loginResponse, response)
	}
}
