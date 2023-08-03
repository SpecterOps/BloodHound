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

package auth

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/specterops/bloodhound/src/model"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var (
	ErrorInvalidOTP = fmt.Errorf("invalid one time password")
)

func GenerateTOTPSecret(issuer, accountName string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
	})
}

func ValidateTOTPSecret(otp string, secret model.AuthSecret) error {
	if !secret.TOTPActivated || totp.Validate(otp, secret.TOTPSecret) {
		return nil
	} else {
		return ErrorInvalidOTP
	}
}

func GenerateQRCodeBase64(key otp.Key) (string, error) {
	if img, err := key.Image(200, 200); err != nil {
		return "", err
	} else {
		var buf bytes.Buffer

		if err := png.Encode(&buf, img); err != nil {
			return "", err
		}

		return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
	}
}
