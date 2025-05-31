// Copyright 2025 Specter Ops, Inc.
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
package license

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func generateLicenseHeader(commentPrefix string) string {
	var (
		// Using a string builder for the formatted header so we can easily return the final string
		formattedHeader strings.Builder
		year            = strconv.Itoa(time.Now().Year())
	)

	// XML and CSS style comments require special rules, this first check handles creating the top of the comment
	if commentPrefix == "<!--" {
		formattedHeader.WriteString("<!--\n")
	} else if commentPrefix == "/*" {
		formattedHeader.WriteString("/*\n")
	}

	for _, line := range strings.Split(licenseHeader, "\n") {
		// We grab the copyright line and edit the year into it inline for efficiency
		if strings.HasPrefix(line, "Copyright") {
			line = strings.ReplaceAll(line, "XXXX", year)
		}

		// XML and CSS style comments should be indented, but not use the prefix
		if commentPrefix == "<!--" || commentPrefix == "/*" {
			formattedHeader.WriteString("    ")
			formattedHeader.WriteString(line)
			formattedHeader.WriteRune('\n')
		} else {
			formattedHeader.WriteString(commentPrefix)
			// only add a space after the comment prefix if the line is non-empty and there's a non-empty prefix
			if len(line) > 0 && len(commentPrefix) > 0 {
				formattedHeader.WriteString(" ")
			}
			formattedHeader.WriteString(line)
			formattedHeader.WriteRune('\n')
		}
	}

	// XML style comments must be properly ended on a new line
	if commentPrefix == "<!--" {
		formattedHeader.WriteString("-->\n")
	} else if commentPrefix == "/*" {
		formattedHeader.WriteString("*/\n")
	}

	return formattedHeader.String()
}

func writeFile(path string, formattedHeaderContent string) error {
	// Get original file info to preserve permissions
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("could not stat file to write %s: %w", path, err)
	}
	originalPerm := fileInfo.Mode().Perm()

	// Open file at path in read/write mode with original permissions
	file, err := os.OpenFile(path, os.O_RDWR, originalPerm)
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", path, err)
	}
	defer file.Close()

	fileReader := bufio.NewReader(file)

	// Create a new temp file to write our updated lines into
	tmpFile, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path))
	if err != nil {
		return fmt.Errorf("could not open temporary file for writing: %w", err)
	}
	// We want to make sure to both close and remove the temporary file when we leave this function, so do it all in one defer
	defer func() {
		if err := tmpFile.Close(); err != nil {
			slog.Error("could not close temporary file", slog.String("err", err.Error()), slog.String("tmpFile", tmpFile.Name()))
		}

		if err := os.Remove(tmpFile.Name()); err != nil {
			slog.Error("could not remove temporary file", slog.String("err", err.Error()), slog.String("tmpFile", tmpFile.Name()))
		}
	}()

	tmpFileWriter := bufio.NewWriter(tmpFile)

	// Get the first line
	firstLine, err := fileReader.ReadString('\n')
	// If there's only one line and it has a valid XML document start in it, we'll bail because we don't support that case
	if errors.Is(err, io.EOF) && strings.Contains(firstLine, "<?xml") {
		return fmt.Errorf("file is single line and has xml, cannot write license for %s", path)
	}
	// If there's only one line, but we've already confirmed it's not a valid XML document, write the header and then the first line
	if errors.Is(err, io.EOF) {
		if _, err := tmpFileWriter.WriteString(formattedHeaderContent); err != nil {
			return fmt.Errorf("could not write formatted header to temp file for path %s: %w", path, err)
		} else if _, err := tmpFileWriter.WriteString(firstLine); err != nil {
			return fmt.Errorf("could not write first line to temp file for path %s: %w", path, err)
		}
	}
	// If there's an unknown error, bail
	if err != nil {
		return fmt.Errorf("could not read first line of file %s: %w", path, err)
	}

	// If we have the start of a valid XML document
	if strings.Contains(firstLine, "<?xml") {
		// And if we can see the end of the valid XML starting tag in the same line
		if strings.Contains(firstLine, "?>") {
			// Then write the first line (the XML document tag) before writing the header
			if _, err := tmpFileWriter.WriteString(firstLine); err != nil {
				return fmt.Errorf("could not write formatted header to temp file for path %s: %w", path, err)
			} else if _, err := tmpFileWriter.WriteString(formattedHeaderContent); err != nil {
				return fmt.Errorf("could not write first line to temp file for path %s: %w", path, err)
			}
		} else {
			// Otherwise, bail out because we may end up writing inside of a multi-line tag
			return fmt.Errorf("could not write license due to unsupported starting xml tag in %s", path)
		}
	} else {
		// In non-XML cases, assume we can safely write our header first, and then append the first line
		if _, err := tmpFileWriter.WriteString(formattedHeaderContent); err != nil {
			return fmt.Errorf("could not write formatted header to temp file for path %s: %w", path, err)
		} else if _, err := tmpFileWriter.WriteString(firstLine); err != nil {
			return fmt.Errorf("could not write first line to temp file for path %s: %w", path, err)
		}
	}

	fileScanner := bufio.NewScanner(fileReader)
	linesBuffered := 1

	// Scan through the rest of the file line-by-line, appending to the tmp file with a newline added back in
	for fileScanner.Scan() {
		if _, err := tmpFileWriter.WriteString(fileScanner.Text() + "\n"); err != nil {
			return fmt.Errorf("could not write remaining file to tmp file for %s: %w", path, err)
		}
		// Flush every 10 lines to keep memory reasonable
		if linesBuffered >= 10 {
			if err := tmpFileWriter.Flush(); err != nil {
				return fmt.Errorf("could not flush lines to file %s: %w", path, err)
			}
			linesBuffered = 0
		} else {
			linesBuffered++
		}
	}
	// Handle any additional scanner errors
	if err := fileScanner.Err(); err != nil {
		return fmt.Errorf("file scan was not successful for %s: %w", path, err)
	}

	// Make sure to flush the buffer to the filesystem before moving on
	if err := tmpFileWriter.Flush(); err != nil {
		return fmt.Errorf("could not flush remaining temp file lines to disk for %s: %w", path, err)
	}

	// We need to truncate the original file and seek back to the start of both the original file and the temp file before continuing
	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("could not truncate file %s: %w", path, err)
	} else if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("could not seek to beginning of file %s: %w", path, err)
	} else if _, err := tmpFile.Seek(0, 0); err != nil {
		return fmt.Errorf("could not seek to beginning of temp file for %s: %w", path, err)
	}

	tmpFileScanner := bufio.NewScanner(tmpFile)
	fileWriter := bufio.NewWriter(file)
	linesBuffered = 1

	// Scan through the entire temp file line-by-line and write each line to the original file path
	for tmpFileScanner.Scan() {
		if _, err := fileWriter.WriteString(tmpFileScanner.Text() + "\n"); err != nil {
			return fmt.Errorf("could not write temp file contents to file %s: %w", path, err)
		}
		// Flush every 10 lines to keep memory reasonable
		if linesBuffered >= 10 {
			if err := fileWriter.Flush(); err != nil {
				return fmt.Errorf("could not flush lines to file %s: %w", path, err)
			}
			linesBuffered = 0
		} else {
			linesBuffered += 1
		}
	}
	// Handle any additional scanner errors
	if err := tmpFileScanner.Err(); err != nil {
		return fmt.Errorf("file scan was not successful for %s: %w", path, err)
	}

	if err := fileWriter.Flush(); err != nil {
		return fmt.Errorf("could not flush final lines to file %s: %w", path, err)
	}

	return nil
}
