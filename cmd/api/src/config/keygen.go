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

package config

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"strings"
)

// generateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomBase64String returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBase64String(n int) (string, error) {
	b, err := generateRandomBytes(n)
	return base64.StdEncoding.EncodeToString(b), err
}

func GenerateSecureRandomString(n int) (string, error) {
	// charset represents characters that can be easily selected together in any terminal
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

	var (
		builder     strings.Builder
		charsetSize = big.NewInt(int64(len(charset)))
	)

	for i := 0; i < n; i++ {
		if charIdx, err := rand.Int(rand.Reader, charsetSize); err != nil {
			return "", err
		} else {
			builder.WriteByte(charset[charIdx.Int64()])
		}
	}

	return builder.String(), nil
}
