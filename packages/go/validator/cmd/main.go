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

	"github.com/specterops/bloodhound/packages/go/validator"
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
		slog.Error("Failed to open file", "fileName", fileName, "error", err)
		os.Exit(1)
	}

	jsonSchema, err := validator.LoadIngestSchema()
	if err != nil {
		slog.Error("Failed to load ingest schema", "error", err)
		os.Exit(1)
	}

	reader.Seek(0, io.SeekStart)

	v := validator.NewValidator(reader, jsonSchema)

	report, err := v.ParseAndValidate()
	if err != nil {
		slog.Error("Failed to validate", "err", err)
	}

	var w io.WriteCloser

	if output != "" {
		file, err := os.Create(output)
		if err != nil {
			slog.Error("Failed to open output file", "err", err)
		}

		w = file
	} else {
		w = os.Stdout
	}

	outputReport(w, report)
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
	return fmt.Sprintf("CRITICAL ERROR:\n%s", e.Message)
}

func formatValidationError(e validator.ValidationError) (string, error) {
	var sb strings.Builder

	sb.WriteString("VALIDATION ERROR:\n")

	sb.WriteString("Location: " + e.Location + "\n")

	var objBytes bytes.Buffer
	err := json.Indent(&objBytes, []byte(e.RawObject), "", "\t")
	if err != nil {
		return "", err
	}

	sb.WriteString("Object:\n" + objBytes.String() + "\n")

	sb.WriteString("Errors:\n")
	for _, e := range e.Errors {
		sb.WriteString("at " + e.Location + ": " + e.Error + "\n")
	}

	return sb.String(), nil
}
