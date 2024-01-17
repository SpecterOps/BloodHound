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
	"testing"

	"golang.org/x/crypto/argon2"
)

func TestArgon2_Digest(t *testing.T) {
	argon2Digester := Argon2{
		MemoryKibibytes: 1024,
		NumIterations:   1,
		NumThreads:      1,
	}

	if _, err := argon2Digester.Digest("This is a test."); err != nil {
		t.Fatalf("Unexpected error while performing argon2 digest: %v", err)
	}
}

func TestArgon2_ParseDigest(t *testing.T) {
	const digestMCFormatString = "$argon2id$v=19$m=1024,t=2,p=2$4TrXa715awDIAbOUi+SGkg==$1cJNBKtLQ2St83ttOv6xVw=="

	argon2Digester := Argon2{}

	if digestInst, err := argon2Digester.ParseDigest(digestMCFormatString); err != nil {
		t.Fatalf("Unexpected error while parsing argon2 digest: %v", err)
	} else {
		digest := digestInst.(Argon2Digest)

		if digest.DigestVariant != Argon2idVariant {
			t.Fatalf("Expected argon2 variant of digest to be %s but got: %s", Argon2idVariant, digest.DigestVariant)
		} else if digest.Version != argon2.Version {
			t.Fatalf("Expected argon2 version of digest to be %d but got: %d", argon2.Version, digest.Version)
		} else if digest.MemoryKibibytes != 1024 {
			t.Fatalf("Expected MemoryKibibytes of digest to be 1024 but got: %d", digest.MemoryKibibytes)
		} else if digest.NumIterations != 2 {
			t.Fatalf("Expected NumIterations of digest to be 2 but got: %d", digest.NumIterations)
		} else if digest.NumThreads != 2 {
			t.Fatalf("Expected NumThreads of digest to be 2 but got: %d", digest.NumThreads)
		}
	}
}

func TestArgon2Digest_String(t *testing.T) {
	const digestMCFormatString = "$argon2id$v=19$m=1024,t=2,p=2$4TrXa715awDIAbOUi+SGkg==$1cJNBKtLQ2St83ttOv6xVw=="

	argon2Digester := Argon2{}

	if digest, err := argon2Digester.ParseDigest(digestMCFormatString); err != nil {
		t.Fatalf("Unexpected error while parsing argon2 digest: %v", err)
	} else if parsedDigestMCFormatString := digest.String(); parsedDigestMCFormatString != digestMCFormatString {
		t.Fatalf("Expected MC comptabile string to be %s but got %s", digestMCFormatString, parsedDigestMCFormatString)
	} else if _, err := argon2Digester.ParseDigest(parsedDigestMCFormatString); err != nil {
		t.Fatalf("Unexpected error while parsing argon2 digest: %v", err)
	}
}

func TestArgon2Digest_Validate(t *testing.T) {
	const digestMCFormatString = "$argon2id$v=19$m=1024,t=2,p=2$4TrXa715awDIAbOUi+SGkg==$1cJNBKtLQ2St83ttOv6xVw=="

	argon2Digester := Argon2{}

	if digest, err := argon2Digester.ParseDigest(digestMCFormatString); err != nil {
		t.Fatalf("Unexpected error while parsing argon2 digest: %v", err)
	} else if valid := digest.Validate("This is a test."); !valid {
		t.Fatalf(`Expected content "This is a test." to validate against digest: %s`, digestMCFormatString)
	} else if valid := digest.Validate("This should fail."); valid {
		t.Fatalf(`Expected content "This should fail." to fail validation against digest: %s`, digestMCFormatString)
	}
}
