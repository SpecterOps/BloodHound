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

package auth_test

import (
	"testing"
	"time"

	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/model"
	"github.com/pquerna/otp/totp"
)

func TestGenerateTOTPSecret(t *testing.T) {
	if key, err := auth.GenerateTOTPSecret("issuer.io", "accountName"); err != nil {
		t.Fatal(err)
	} else if key.Issuer() != "issuer.io" {
		t.Errorf("got %v, want %v", key.Issuer(), "issuer.io")
	}
}

func TestValidateTOTPSecret(t *testing.T) {
	key, err := auth.GenerateTOTPSecret("issuer.io", "accountName")
	if err != nil {
		t.Fatal(err)
	}

	secret := model.AuthSecret{
		TOTPActivated: false,
	}

	mfaSecret := model.AuthSecret{
		TOTPActivated: true,
		TOTPSecret:    key.Secret(),
	}

	code, err := totp.GenerateCode(key.Secret(), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	type Input struct {
		Code   string
		Secret model.AuthSecret
	}

	cases := []struct {
		Input Input
		Error error
	}{
		{Input{code, mfaSecret}, nil},
		{Input{"", secret}, nil},
		{Input{"", mfaSecret}, auth.ErrorInvalidOTP},
	}

	for _, tc := range cases {
		if err := auth.ValidateTOTPSecret(tc.Input.Code, tc.Input.Secret); err != nil {
			if tc.Error == nil {
				t.Errorf("got %v, want %v", err, tc.Error)
			} else if err.Error() != tc.Error.Error() {
				t.Errorf("got %v, want %v", err, tc.Error)
			}
		} else if tc.Error != nil {
			t.Errorf("got %v, want %v", err, tc.Error)
		}
	}
}

func TestGenerateQRCodeBase64(t *testing.T) {

	key, err := auth.GenerateTOTPSecret("issuer.io", "accountName")
	if err != nil {
		t.Fatal(err)
	}

	if qr, err := auth.GenerateQRCodeBase64(*key); err != nil {
		t.Errorf("got %v, want %v", err, nil)
	} else if qr == "" {
		t.Error("expected qr to not be empty")
	}
}
