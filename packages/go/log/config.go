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

package log

import (
	"time"

	"github.com/rs/zerolog"
)

const (
	// TimeFormatUnix defines a time format that makes time fields to be serialized as Unix timestamp integers.
	TimeFormatUnix = ""

	// TimeFormatUnixMs defines a time format that makes time fields to be serialized as Unix timestamp integers in
	// milliseconds.
	TimeFormatUnixMs = "UNIXMS"

	// TimeFormatUnixMicro defines a time format that makes time fields to be serialized as Unix timestamp integers in
	// microseconds.
	TimeFormatUnixMicro = "UNIXMICRO"

	// TimeFormatUnixNano defines a time format that makes time fields to be serialized as Unix timestamp integers in
	// nanoseconds.
	TimeFormatUnixNano = "UNIXNANO"
)

type Configuration struct {
	// Level defines the global logging verbosity level.
	Level Level

	// CallerSkipFrameCount is the number of stack frames to skip to find the caller.
	CallerSkipFrameCount int

	// DurationFieldUnit defines the unit for time.Duration type fields added using the Dur function.
	DurationFieldUnit time.Duration

	// TimeFieldFormat defines the time format of the Time field type. If set to TimeFormatUnix, TimeFormatUnixMs,
	// TimeFormatUnixMicro or TimeFormatUnixNano, the time is formatted as a UNIX timestamp as integer.
	TimeFieldFormat string
}

// WithLevel returns this configuration with the specified log verbosity level.
func (s *Configuration) WithLevel(level Level) *Configuration {
	s.Level = level
	return s
}

// WithCallerSkipFrameCount returns this configuration with the specified CallerSkipFrameCount.
func (s *Configuration) WithCallerSkipFrameCount(callerSkipFrameCount int) *Configuration {
	s.CallerSkipFrameCount = callerSkipFrameCount
	return s
}

// WithDurationFieldUnit returns this configuration with the specified DurationFieldUnit.
func (s *Configuration) WithDurationFieldUnit(durationFieldUnit time.Duration) *Configuration {
	s.DurationFieldUnit = durationFieldUnit
	return s
}

// WithTimeFieldFormat returns this configuration with the specified TimeFieldFormat.
func (s *Configuration) WithTimeFieldFormat(timeFieldFormat string) *Configuration {
	s.TimeFieldFormat = timeFieldFormat
	return s
}

// DefaultConfiguration returns a set of default configuration values for logging.
func DefaultConfiguration() *Configuration {
	return &Configuration{
		Level:                LevelInfo,
		CallerSkipFrameCount: 2,
		DurationFieldUnit:    time.Millisecond,
		TimeFieldFormat:      time.RFC3339Nano,
	}
}

// Configure applies the given configuration to the logging framework.
func Configure(config *Configuration) {
	zerolog.TimeFieldFormat = config.TimeFieldFormat
	zerolog.CallerSkipFrameCount = config.CallerSkipFrameCount
	zerolog.DurationFieldUnit = config.DurationFieldUnit

	SetGlobalLevel(config.Level)
}

// ConfigureDefaults is a helper function that configures the logging framework with a set of reasonable default
// settings.
func ConfigureDefaults() {
	Configure(DefaultConfiguration())
}
