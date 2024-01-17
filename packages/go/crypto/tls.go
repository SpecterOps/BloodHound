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

package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"
)

const (
	DefaultRSABitSize    = 4096
	PEMTypeCertificate   = "CERTIFICATE"
	PEMTypeRSAPrivateKey = "RSA PRIVATE KEY"
)

func tryParsePrivateKey(key string) (*rsa.PrivateKey, error) {
	keyBlock, _ := pem.Decode([]byte(key))

	if pkcs8PrivateKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes); err == nil {
		if rsaPrivateKey, ok := pkcs8PrivateKey.(*rsa.PrivateKey); !ok {
			return nil, fmt.Errorf("unsupported key type")
		} else {
			return rsaPrivateKey, nil
		}
	}

	return x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
}

func X509ParseCert(cert string) (*x509.Certificate, error) {
	formattedCert := cert

	if !strings.HasPrefix("-----BEGIN CERTIFICATE-----", formattedCert) {
		formattedCert = "-----BEGIN CERTIFICATE-----\n" + formattedCert
	}

	if !strings.HasSuffix("-----END CERTIFICATE----- ", formattedCert) {
		formattedCert = formattedCert + "\n-----END CERTIFICATE----- "
	}

	if certBlock, _ := pem.Decode([]byte(formattedCert)); certBlock == nil {
		return nil, fmt.Errorf("unable to decode cert")
	} else if cert, err := x509.ParseCertificate(certBlock.Bytes); err != nil {
		return nil, err
	} else {
		return cert, nil
	}
}

func X509ParsePair(cert, key string) (*x509.Certificate, *rsa.PrivateKey, error) {
	formattedCert := cert

	if !strings.HasPrefix("-----BEGIN CERTIFICATE-----", formattedCert) {
		formattedCert = "-----BEGIN CERTIFICATE-----\n" + formattedCert
	}

	if !strings.HasSuffix("-----END CERTIFICATE----- ", formattedCert) {
		formattedCert = formattedCert + "\n-----END CERTIFICATE----- "
	}

	if certBlock, _ := pem.Decode([]byte(formattedCert)); certBlock == nil {
		return nil, nil, fmt.Errorf("unable to decode cert")
	} else if cert, err := x509.ParseCertificate(certBlock.Bytes); err != nil {
		return nil, nil, err
	} else if key, err := tryParsePrivateKey(key); err != nil {
		return nil, nil, err
	} else {
		return cert, key, err
	}
}

func X509StringPair(cert *x509.Certificate, privateKey *rsa.PrivateKey) (string, string, error) {
	var (
		certBlock = &pem.Block{
			Type:  PEMTypeCertificate,
			Bytes: cert.Raw,
		}

		keyBlock = &pem.Block{
			Type:  PEMTypeRSAPrivateKey,
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}

		buffer = &bytes.Buffer{}
	)

	if err := pem.Encode(buffer, certBlock); err != nil {
		return "", "", err
	}

	certPEM := buffer.String()
	buffer.Reset()

	if err := pem.Encode(buffer, keyBlock); err != nil {
		return "", "", err
	}

	return certPEM, buffer.String(), nil
}

func X509WritePair(certDERBytes []byte, privateKey *rsa.PrivateKey, tlsCertFile, tlsKeyFile string) error {
	if fout, err := os.OpenFile(tlsCertFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644); err != nil {
		return err
	} else {
		defer fout.Close()

		block := &pem.Block{
			Type:  PEMTypeCertificate,
			Bytes: certDERBytes,
		}

		if err := pem.Encode(fout, block); err != nil {
			return err
		}
	}

	if fout, err := os.OpenFile(tlsKeyFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644); err != nil {
		return err
	} else {
		defer fout.Close()

		block := &pem.Block{
			Type:  PEMTypeRSAPrivateKey,
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}

		if err := pem.Encode(fout, block); err != nil {
			return err
		}
	}

	return nil
}

func TLSReadPairFiles(certFile, keyFile string) (tls.Certificate, error) {
	if certPEMContent, err := os.ReadFile(certFile); err != nil {
		return tls.Certificate{}, err
	} else if keyPEMContent, err := os.ReadFile(keyFile); err != nil {
		return tls.Certificate{}, err
	} else {
		return tls.X509KeyPair(certPEMContent, keyPEMContent)
	}
}

func TLSGeneratePair(organization string) (*x509.Certificate, *rsa.PrivateKey, error) {
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(130), nil).Sub(max, big.NewInt(1))

	if privateKey, err := rsa.GenerateKey(rand.Reader, DefaultRSABitSize); err != nil {
		return nil, nil, err
	} else if serial, err := rand.Int(rand.Reader, max); err != nil {
		return nil, nil, err
	} else {

		template := x509.Certificate{
			SerialNumber: serial,
			Subject: pkix.Name{
				Organization: []string{organization},
			},

			NotBefore: time.Now(),
			NotAfter:  time.Now().Add(time.Hour * 24 * 180),

			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}

		if derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey); err != nil {
			return nil, nil, err
		} else if certificate, err := x509.ParseCertificate(derBytes); err != nil {
			return nil, nil, err
		} else {
			return certificate, privateKey, nil
		}
	}
}

func TLSGenerateStringPair(organization string) (string, string, error) {
	if cert, key, err := TLSGeneratePair(organization); err != nil {
		return "", "", err
	} else {
		return X509StringPair(cert, key)
	}
}
