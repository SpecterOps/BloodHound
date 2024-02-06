// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/specterops/bloodhound/headers"
	"github.com/stretchr/testify/require"
)

func TestGetScheme(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "foo/bar", nil)
	require.Nil(t, err)
	require.Equal(t, "http", getScheme(req))

	secureReq, err := http.NewRequestWithContext(ctx, "GET", "foo/bar", nil)
	require.Nil(t, err)

	secureReq.TLS = &tls.ConnectionState{}
	require.Equal(t, "https", getScheme(secureReq))

	protoReq, err := http.NewRequestWithContext(ctx, "GET", "foo/bar", nil)
	require.Nil(t, err)
	q := url.Values{}
	protoReq.Header.Set("X-Forwarded-Proto", "foobar")
	protoReq.URL.RawQuery = q.Encode()
	require.Equal(t, "foobar", getScheme(protoReq))
}

func TestRequestWaitDuration_Failure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "foo/bar", nil)
	require.Nil(t, err)

	q := url.Values{}
	req.Header.Set(headers.Prefer.String(), "wait=1.5")
	req.URL.RawQuery = q.Encode()

	_, err = requestWaitDuration(req, 30*time.Second)
	require.NotNil(t, err)
}

func TestRequestWaitDuration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "foo/bar", nil)
	require.Nil(t, err)

	q := url.Values{}
	req.Header.Set(headers.Prefer.String(), "wait=1")
	req.URL.RawQuery = q.Encode()

	requestedWaitDuration, err := requestWaitDuration(req, 30*time.Second)
	require.Nil(t, err)
	require.Equal(t, 1*time.Second, requestedWaitDuration.Value)
	require.True(t, requestedWaitDuration.UserSet)
}

func TestParseUserIP_XForwardedFor_RemoteAddr(t *testing.T) {
	req, err := http.NewRequest("GET", "/teapot", nil)
	require.Nil(t, err)

	ip1 := "192.168.1.1:8080"
	ip2 := "192.168.1.2"
	ip3 := "192.168.1.3"

	req.Header.Set("X-Forwarded-For", strings.Join([]string{ip1, ip2, ip3}, ","))
	req.RemoteAddr = "0.0.0.0:3000"

	require.Equal(t, parseUserIP(req), strings.Join([]string{ip1, ip2, ip3, "0.0.0.0"}, ","))
}

func TestParseUserIP_RemoteAddrOnly(t *testing.T) {
	req, err := http.NewRequest("GET", "/teapot", nil)
	require.Nil(t, err)
	req.RemoteAddr = "0.0.0.0:3000"
	require.Equal(t, parseUserIP(req), "0.0.0.0")
}

func TestParsePreferHeaderWait(t *testing.T) {
	_, err := parsePreferHeaderWait("wait=1.5", 30*time.Second)
	require.NotNil(t, err)

	duration, err := parsePreferHeaderWait("wait=5", 30*time.Second)
	require.Nil(t, err)
	require.Equal(t, 5*time.Second, duration)

	duration, err = parsePreferHeaderWait("", 30*time.Second)
	require.Nil(t, err)
	require.Equal(t, 30*time.Second, duration)
}
