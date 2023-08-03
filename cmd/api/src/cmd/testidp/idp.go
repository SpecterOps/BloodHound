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

package main

// Borrowed with love from https://github.com/crewjam/saml/blob/master/example/idp/idp.go
//
// License information for crewjam can be found in samlidp/LICENSE

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/specterops/bloodhound/crypto"
	samlidp2 "github.com/specterops/bloodhound/src/cmd/testidp/samlidp"

	"github.com/crewjam/saml/logger"
	"github.com/zenazn/goji"
	"golang.org/x/crypto/bcrypt"
)

func createDefaultUsers(organization string, store samlidp2.Store) error {
	if hashedPassword, err := bcrypt.GenerateFromPassword([]byte("hunter2"), bcrypt.DefaultCost); err != nil {
		return err
	} else {
		var (
			alice = samlidp2.User{
				Name:           "alice",
				HashedPassword: hashedPassword,
				Groups:         []string{"Administrators", "Users"},
				Email:          fmt.Sprintf("alice@%s", organization),
				CommonName:     "Alice Smith",
				Surname:        "Smith",
				GivenName:      "Alice",
			}

			bob = samlidp2.User{
				Name:           "bob",
				HashedPassword: hashedPassword,
				Groups:         []string{"Users"},
				Email:          fmt.Sprintf("bob@%s", organization),
				CommonName:     "Bob Smith",
				Surname:        "Smith",
				GivenName:      "Bob",
			}
		)

		if err := store.Put("/users/alice", alice); err != nil {
			return err
		}

		if err := store.Put("/users/bob", bob); err != nil {
			return err
		}
	}

	return nil
}

func startIDP(baseURL *url.URL, organization string, cert *x509.Certificate, key *rsa.PrivateKey) {
	idpOptions := samlidp2.Options{
		URL:              *baseURL,
		Key:              key,
		Logger:           logger.DefaultLogger,
		Certificate:      cert,
		Store:            &samlidp2.MemoryStore{},
		AllowUntrustedSP: true,
	}

	if idpServer, err := samlidp2.New(idpOptions); err != nil {
		fmt.Printf("failed to create IDP instance: %v\n", err)
	} else if err := createDefaultUsers(organization, idpServer.Store); err != nil {
		fmt.Printf("failed to create default users: %v\n", err)
	} else {
		goji.Handle("/*", idpServer)
		goji.Serve()

		//goji.ServeTLS(&tls.Config{
		//	Certificates: []tls.Certificate{{
		//		Certificate: [][]byte{
		//			cert.Raw,
		//		},
		//		PrivateKey: key,
		//	}},
		//})
	}
}

func writePEMBlockToFile(path string, block *pem.Block) error {
	if fout, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644); err != nil {
		return err
	} else {
		defer fout.Close()
		return pem.Encode(fout, block)
	}
}

func generateNewCerts(keyPath, certPath, organization string) (*x509.Certificate, *rsa.PrivateKey, error) {
	if cert, key, err := crypto.TLSGeneratePair(organization); err != nil {
		return nil, nil, err
	} else {
		var (
			keyPEMBlock = &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(key),
			}
			certPEMBlock = &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: cert.Raw,
			}
		)

		if err := writePEMBlockToFile(keyPath, keyPEMBlock); err != nil {
			return nil, nil, err
		}

		if err := writePEMBlockToFile(certPath, certPEMBlock); err != nil {
			return nil, nil, err
		}

		return cert, key, nil
	}
}

func getIDPCerts(keyPath, certPath, organization string) (*x509.Certificate, *rsa.PrivateKey, error) {
	if keyBytes, err := os.ReadFile(keyPath); err != nil {
		if os.IsNotExist(err) {
			return generateNewCerts(keyPath, certPath, organization)
		}

		return nil, nil, err
	} else if certBytes, err := os.ReadFile(certPath); err != nil {
		if os.IsNotExist(err) {
			return generateNewCerts(keyPath, certPath, organization)
		}

		return nil, nil, err
	} else {
		return crypto.X509ParsePair(string(certBytes), string(keyBytes))
	}
}

func main() {
	var (
		keyPath      string
		certPath     string
		rawBaseURL   string
		organization string
	)

	flag.StringVar(&certPath, "cert", "", "Path to the IDP cert. If the file doesn't exist a new cert and keypair will be created.")
	flag.StringVar(&keyPath, "key", "", "Path to the IDP key. If the file doesn't exist a new cert and keypair will be created.")
	flag.StringVar(&rawBaseURL, "url", "", "The base URL to the IDP.")
	flag.StringVar(&organization, "org", "example.com", "The organization of the IDP.")
	flag.Parse()

	if rawBaseURL == "" {
		fmt.Printf("The IDP base URL is a required parameter\n")
		flag.Usage()
	} else if baseURL, err := url.Parse(rawBaseURL); err != nil {
		fmt.Printf("Can not parse base URL: %v\n", err)
	} else if cert, key, err := getIDPCerts(keyPath, certPath, organization); err != nil {
		fmt.Printf("Unable to generate self-signed certs: %v\n", err)
	} else {
		startIDP(baseURL, organization, cert, key)
	}
}
