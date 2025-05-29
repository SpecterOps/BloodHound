// Copyright 2025 Specter Ops, Inc.
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

package tls

import (
	"crypto/rsa"
	"crypto/x509"

	"github.com/specterops/bloodhound/crypto"
	"github.com/specterops/bloodhound/src/config"
)

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/tls.go -package=mocks . Service

// Serves as a lightweight wrapper around BloodHound's crypto package.
type Service interface {
	Parse(cfg config.SAMLConfiguration) (*x509.Certificate, *rsa.PrivateKey, error)
}

type Client struct{}

// Parse allows for the certification parsing to be abstracted.
func (c *Client) Parse(cfg config.SAMLConfiguration) (*x509.Certificate, *rsa.PrivateKey, error) {
	return crypto.X509ParsePair(cfg.ServiceProviderCertificate, cfg.ServiceProviderKey)
}
