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

package upload

import (
	"io"

	"github.com/specterops/bloodhound/bomenc"
	"github.com/specterops/bloodhound/src/model/ingest"
)

// FileValidator defines the interface for ingest file validation.
// It receives a source reader (src) and a destination writer (dst).
// Implementations are responsible for validating the input stream,
// while simultaneously copying it to the destination for persistence.
// This abstraction supports format-agnostic payloads (e.g., JSON, ZIP)
type FileValidator func(src io.Reader, dst io.Writer) (ingest.Metadata, error)

// WriteAndValidateZIP implements FileValidator for ZIP ingest files.
func WriteAndValidateZip(src io.Reader, dst io.Writer) (ingest.Metadata, error) {
	tr := io.TeeReader(src, dst)
	return ingest.Metadata{}, ValidateZipFile(tr)
}

// IngestValidator encapsulates precompiled JSON schemas used to validate
// graph ingest payloads, including node and edge definitions.
//
// This struct allows schema compilation to happen once at application startup,
// avoiding repeated compilation during each file ingest request.
type IngestValidator struct {
	IngestSchema IngestSchema
}

func NewIngestValidator(schema IngestSchema) IngestValidator {
	return IngestValidator{
		IngestSchema: schema,
	}
}

// WriteAndValidateJSON implements FileValidator for JSON ingest files.
// It streams JSON through a validator while simultaneously writing it to disk.
//
// This method is a member of `IngestValidator` to reuse the precompiled
// node and edge schemas loaded during application startup.
func (s *IngestValidator) WriteAndValidateJSON(src io.Reader, dst io.Writer) (ingest.Metadata, error) {
	normalizedContent, err := bomenc.NormalizeToUTF8(src)
	if err != nil {
		return ingest.Metadata{}, err
	}
	tr := io.TeeReader(normalizedContent, dst)
	metatag, err := ParseAndValidatePayload(tr, s.IngestSchema, true, true)

	return metatag, err
}
