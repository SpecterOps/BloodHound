// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	validator "github.com/specterops/bloodhound/packages/go/chow/ingestvalidator"
)

var (
	output string
)

func main() {
	flag.StringVar(&output, "output", "", "output file")
	flag.Parse()

	files := flag.Args()

	if len(files) < 1 {
		slog.Error("No files provided")
		os.Exit(1)
	}

	fileName := files[0]

	reader, err := os.Open(fileName)
	if err != nil {
		slog.Error("Failed to open file",
			slog.String("file_name", fileName),
			attr.Error(err),
		)
		os.Exit(1)
	}
	defer reader.Close()

	jsonSchema, err := validator.LoadIngestSchema()
	if err != nil {
		slog.Error("Failed to load ingest schema", attr.Error(err))
		os.Exit(1)
	}

	v := validator.NewValidator(reader, jsonSchema)

	_, report, err := v.ParseAndValidate()
	validationFailed := err != nil
	if validationFailed {
		slog.Error("Validation failed", attr.Error(err))
	}

	var w io.WriteCloser

	if output != "" {
		file, err := os.Create(output)
		if err != nil {
			slog.Error("Failed to open output file", attr.Error(err))
			os.Exit(1)
		}
		defer file.Close()

		w = file
	} else {
		w = os.Stdout
	}

	outputReport(w, report)

	if validationFailed {
		os.Exit(1)
	}
}

func outputReport(w io.WriteCloser, report validator.ValidationReport) error {
	for _, e := range report.CriticalErrors {
		_, err := w.Write([]byte(formatCriticalError(e)))
		if err != nil {
			return err
		}

		_, err = w.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}

	for _, e := range report.ValidationErrors {
		s, err := formatValidationError(e)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(s))
		if err != nil {
			return err
		}

		_, err = w.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}

	return nil
}

func formatCriticalError(e validator.CriticalError) string {
	return fmt.Sprintf("CRITICAL ERROR:\n%s\n%v", e.Message, e.Error)
}

func formatValidationError(valErr validator.ValidationError) (string, error) {
	var (
		sb       strings.Builder
		objBytes bytes.Buffer
	)

	sb.WriteString("VALIDATION ERROR:\n")

	sb.WriteString("Location: " + valErr.Location + "\n")

	err := json.Indent(&objBytes, []byte(valErr.RawObject), "", "\t")
	if err != nil {
		return "", err
	}

	sb.WriteString("Object:\n" + objBytes.String() + "\n")

	sb.WriteString("Errors:\n")
	for _, e := range valErr.Errors {
		sb.WriteString("at " + e.Location + ": " + e.Error + "\n")
	}

	return sb.String(), nil
}
