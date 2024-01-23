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
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Level is a type alias that represents a log verbosity level.
type Level = zerolog.Level

const (
	// LevelPanic is the least verbose log level. This should be reserved for any logging statements that MUST be
	// written regardless of the configured global or event logging level.
	LevelPanic = zerolog.PanicLevel

	// LevelFatal is on of the lease verbose log levels reserved for error messages that would result in the
	// application becoming non-functional.
	LevelFatal = zerolog.FatalLevel

	// LevelError is a log verbosity level reserved for error reporting.
	LevelError = zerolog.ErrorLevel

	// LevelWarn is a log verbosity level reserved for warnings that are non-critical but otherwise noteworthy to
	// consumers of log data.
	LevelWarn = zerolog.WarnLevel

	// LevelInfo is a log verbosity level reserved for information that is non-critical but otherwise noteworthy to
	// consumers of log data.
	LevelInfo = zerolog.InfoLevel

	// LevelDebug is a log verbosity level reserved for nosier information that is non-critical but essential to
	// debugging application workflows.
	LevelDebug = zerolog.DebugLevel

	// LevelTrace is the most verbose log level. This should be reserved for any logging statements that would
	// otherwise flood the application log with data of questionable value.
	LevelTrace = zerolog.TraceLevel

	// LevelNone defines an absent log verbosity level.
	LevelNone = zerolog.NoLevel

	// LevelDisabled defines a log verbosity level that disables the sending of logging events.
	LevelDisabled = zerolog.Disabled

	FieldElapsed       = "elapsed"
	FieldMeasurementID = "measurement_id"
)

// The values below are configured in this manner to avoid having to rely on default values that may change throughout
// the development lifecycle of zerolog. While inefficient this is only done once at time of import of the log module.
var (
	levelPanicValue    = strings.ToLower(zerolog.PanicLevel.String())
	levelFatalValue    = strings.ToLower(zerolog.FatalLevel.String())
	levelErrorValue    = strings.ToLower(zerolog.ErrorLevel.String())
	levelWarnValue     = strings.ToLower(zerolog.WarnLevel.String())
	levelInfoValue     = strings.ToLower(zerolog.InfoLevel.String())
	levelDebugValue    = strings.ToLower(zerolog.DebugLevel.String())
	levelTraceValue    = strings.ToLower(zerolog.TraceLevel.String())
	levelNoneValue     = strings.ToLower(zerolog.NoLevel.String())
	levelDisabledValue = strings.ToLower(zerolog.Disabled.String())
)

// ParseLevel takes a string value and attempts to match it to the string representation of one of the log verbosity
// levels available. This defaults to LevelNone along with an error if a match can not be found.
func ParseLevel(rawLevel string) (Level, error) {
	switch strings.ToLower(rawLevel) {
	case levelPanicValue:
		return LevelPanic, nil

	case levelFatalValue:
		return LevelFatal, nil

	case levelErrorValue:
		return LevelError, nil

	case levelWarnValue:
		return LevelWarn, nil

	case levelInfoValue:
		return LevelInfo, nil

	case levelDebugValue:
		return LevelDebug, nil

	case levelTraceValue:
		return LevelTrace, nil

	case levelNoneValue:
		return LevelNone, nil

	case levelDisabledValue:
		return LevelDisabled, nil

	default:
		return LevelNone, fmt.Errorf("unknown log level: %s", rawLevel)
	}
}

// SetGlobalLevel sets the global log verbosity level.
func SetGlobalLevel(level Level) {
	zerolog.SetGlobalLevel(level)
}

// GlobalAccepts returns true if the given log verbosity level would be emitted based on the current global logging
// level.
func GlobalAccepts(level Level) bool {
	return GlobalLevel() <= level
}

// GlobalLevel returns the current global log verbosity level.
func GlobalLevel() Level {
	return zerolog.GlobalLevel()
}

// WithLevel returns a logging event with the given log verbosity level.
func WithLevel(level Level) Event {
	return &event{
		event: log.WithLevel(level),
	}
}

// Panic returns a logging event with the LevelPanic log verbosity level.
func Panic() Event {
	return WithLevel(LevelPanic)
}

// Panicf is a convenience function for writing a log event with the given format and arguments with the LevelPanic
// log verbosity level.
func Panicf(format string, args ...any) {
	Panic().Msgf(format, args...)
}

// Fatalf is a convenience function for writing a log event with the given format and arguments with the LevelFatal
// log verbosity level.
func Fatalf(format string, args ...any) {
	WithLevel(LevelFatal).Msgf(format, args...)
	os.Exit(1)
}

// Error returns a logging event with the LevelError log verbosity level.
func Error() Event {
	return WithLevel(LevelError)
}

// Errorf is a convenience function for writing a log event with the given format and arguments with the LevelError
// log verbosity level.
func Errorf(format string, args ...any) {
	Error().Msgf(format, args...)
}

// Warn returns a logging event with the LevelWarn log verbosity level.
func Warn() Event {
	return WithLevel(LevelWarn)
}

// Warnf is a convenience function for writing a log event with the given format and arguments with the LevelWarn
// log verbosity level.
func Warnf(format string, args ...any) {
	Warn().Msgf(format, args...)
}

// Info returns a logging event with the LevelInfo log verbosity level.
func Info() Event {
	return WithLevel(LevelInfo)
}

// Infof is a convenience function for writing a log event with the given format and arguments with the LevelInfo
// log verbosity level.
func Infof(format string, args ...any) {
	Info().Msgf(format, args...)
}

// Debug returns a logging event with the LevelDebug log verbosity level.
func Debug() Event {
	return WithLevel(LevelDebug)
}

// Debugf is a convenience function for writing a log event with the given format and arguments with the LevelDebug
// log verbosity level.
func Debugf(format string, args ...any) {
	Debug().Msgf(format, args...)
}

// Trace returns a logging event with the LevelTrace log verbosity level.
func Trace() Event {
	return WithLevel(LevelTrace)
}

// Tracef is a convenience function for writing a log event with the given format and arguments with the LevelTrace
// log verbosity level.
func Tracef(format string, args ...any) {
	Trace().Msgf(format, args...)
}

// Measure is a convenience function that returns a deferrable function that will add a runtime duration to the log
// event. The time measurement begins on instantiation of the returned deferrable function and ends upon call of said
// function.
func Measure(level Level, format string, args ...any) func() {
	then := time.Now()

	return func() {
		if elapsed := time.Since(then); elapsed >= measureThreshold {
			WithLevel(level).Duration(FieldElapsed, elapsed).Msgf(format, args...)
		}
	}
}

var (
	logMeasurePairCounter = atomic.Uint64{}
	measureThreshold      = time.Second
)

func SetMeasureThreshold(newMeasureThreshold time.Duration) {
	measureThreshold = newMeasureThreshold
}

func LogAndMeasure(level Level, format string, args ...any) func() {
	var (
		pairID  = logMeasurePairCounter.Add(1)
		message = fmt.Sprintf(format, args...)
		then    = time.Now()
	)

	// Only output the message header on debug
	WithLevel(LevelDebug).Uint64(FieldMeasurementID, pairID).Msg(message)

	return func() {
		if elapsed := time.Since(then); elapsed >= measureThreshold {
			WithLevel(level).Duration(FieldElapsed, elapsed).Uint64(FieldMeasurementID, pairID).Msg(message)
		}
	}
}
