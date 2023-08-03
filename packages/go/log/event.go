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

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../LICENSE.header -destination=./mocks/event.go -package=mocks . Event

import (
	"net"
	"time"

	"github.com/rs/zerolog"
)

// Event represents a log event. It is instanced by one of the level functions and finalized by the Msg or Msgf method.
//
// NOTICE: once finalized the Event should be disposed. Calling any function that performs finalization more than once
// can have unexpected results.
type Event interface {
	// Enabled return false if the Event is going to be filtered out by
	// log level or sampling.
	Enabled() bool

	// Discard disables the event so Msg(f) won't print it.
	Discard() Event

	// Msg sends the Event with msg added as the message field if not empty.
	//
	// NOTICE: once this method is called, the Event should be disposed.
	// Calling Msg twice can have unexpected result.
	Msg(msg string)

	// Send is equivalent to calling Msg("").
	//
	// NOTICE: once this method is called, the Event should be disposed.
	Send()

	// Msgf sends the event with formatted msg added as the message field if not empty.
	//
	// NOTICE: once this method is called, the Event should be disposed.
	// Calling Msgf twice can have unexpected result.
	Msgf(format string, v ...interface{})

	// Str adds the field key with val as a string or []string to the Event context.
	Str(key string, val ...string) Event

	// Bytes adds the field key with val as a string to the Event context.
	//
	// Runes outside of normal ASCII ranges will be hex-encoded in the resulting
	// JSON.
	Bytes(key string, val []byte) Event

	// Hex adds the field key with val as a hex string to the Event context.
	Hex(key string, val []byte) Event

	// Fault adds the field "error" with serialized error or []error to the Event context.
	// If err is nil, no field is added.
	Fault(err ...error) Event

	// Stack enables stack trace printing for any errors passed to Fault(...).
	Stack() Event

	// Bool adds the field key with val as a bool or []bool to the Event context.
	Bool(key string, b ...bool) Event

	// Int adds the field key with i as an int or []int to the Event context.
	Int(key string, i ...int) Event

	// Int8 adds the field key with i as an int8 or []int8 to the Event context.
	Int8(key string, i ...int8) Event

	// Int16 adds the field key with i as an int16 or []int16 to the Event context.
	Int16(key string, i ...int16) Event

	// Int32 adds the field key with i as an int32 or []int32 to the Event context.
	Int32(key string, i ...int32) Event

	// Int64 adds the field key with i as an int64 or []int64 to the Event context.
	Int64(key string, i ...int64) Event

	// Uint adds the field key with i as an uint or []uint to the Event context.
	Uint(key string, i ...uint) Event

	// Uint8 adds the field key with i as an uint8 or []uint8 to the Event context.
	Uint8(key string, i ...uint8) Event

	// Uint16 adds the field key with i as an uint16 or []uint16 to the Event context.
	Uint16(key string, i ...uint16) Event

	// Uint32 adds the field key with i as an uint32 or []uint32 to the Event context.
	Uint32(key string, i ...uint32) Event

	// Uint64 adds the field key with i as an uint64 or []uint64 to the Event context.
	Uint64(key string, i ...uint64) Event

	// Float32 adds the field key with f as a float32 or []float32 to the Event context.
	Float32(key string, f ...float32) Event

	// Float64 adds the field key with f as a float64 or []float64 to the Event context.
	Float64(key string, f ...float64) Event

	// Timestamp adds the current local time as UNIX timestamp to the Event context with the "time" key. This is only
	// useful when setting the time of an event to be submitted in the future.
	//
	// NOTE: This call won't dedupe the "time" key if the Event has one already.
	Timestamp() Event

	// Time adds the field key with t formatted as a string or []string.
	Time(key string, t ...time.Time) Event

	// Duration adds the field key as a duration or []duration.
	Duration(key string, d ...time.Duration) Event

	// TimeDiff adds the field key with positive duration between time t and start.
	// If time t is not greater than start, duration will be 0.
	//
	// Duration format follows the same principle as Dur().
	TimeDiff(key string, t time.Time, start time.Time) Event

	// Any adds the field key with the passed interface marshaled using reflection.
	Any(key string, i any) Event

	// Caller adds the file:line of the caller.
	// The argument skip is the number of stack frames to ascend.
	Caller(skip ...int) Event

	// IPAddr adds IPv4 or IPv6 Address to the event.
	IPAddr(key string, ip net.IP) Event

	// IPPrefix adds IPv4 or IPv6 Prefix (address and mask) to the event.
	IPPrefix(key string, pfx net.IPNet) Event

	// MACAddr adds MAC address to the event.
	MACAddr(key string, ha net.HardwareAddr) Event
}

type event struct {
	event *zerolog.Event
}

func (s *event) Str(key string, val ...string) Event {
	if len(val) > 1 {
		s.event = s.event.Strs(key, val)
	} else {
		s.event = s.event.Str(key, val[0])
	}

	return s
}

func (s *event) Fault(err ...error) Event {
	if len(err) > 1 {
		s.event = s.event.Errs(zerolog.ErrorFieldName, err)
	} else {
		s.event = s.event.Err(err[0])
	}

	return s
}

func (s *event) Bool(key string, b ...bool) Event {
	if len(b) > 1 {
		s.event = s.event.Bools(key, b)
	} else {
		s.event = s.event.Bool(key, b[0])
	}

	return s
}

func (s *event) Int(key string, i ...int) Event {
	if len(i) > 1 {
		s.event = s.event.Ints(key, i)
	} else {
		s.event = s.event.Int(key, i[0])
	}

	return s
}

func (s *event) Int8(key string, i ...int8) Event {
	if len(i) > 1 {
		s.event = s.event.Ints8(key, i)
	} else {
		s.event = s.event.Int8(key, i[0])
	}

	return s
}

func (s *event) Int16(key string, i ...int16) Event {
	if len(i) > 1 {
		s.event = s.event.Ints16(key, i)
	} else {
		s.event = s.event.Int16(key, i[0])
	}

	return s
}

func (s *event) Int32(key string, i ...int32) Event {
	if len(i) > 1 {
		s.event = s.event.Ints32(key, i)
	} else {
		s.event = s.event.Int32(key, i[0])
	}

	return s
}

func (s *event) Int64(key string, i ...int64) Event {
	if len(i) > 1 {
		s.event = s.event.Ints64(key, i)
	} else {
		s.event = s.event.Int64(key, i[0])
	}

	return s
}

func (s *event) Uint(key string, i ...uint) Event {
	if len(i) > 1 {
		s.event = s.event.Uints(key, i)
	} else {
		s.event = s.event.Uint(key, i[0])
	}

	return s
}

func (s *event) Uint8(key string, i ...uint8) Event {
	if len(i) > 1 {
		s.event = s.event.Uints8(key, i)
	} else {
		s.event = s.event.Uint8(key, i[0])
	}

	return s
}

func (s *event) Uint16(key string, i ...uint16) Event {
	if len(i) > 1 {
		s.event = s.event.Uints16(key, i)
	} else {
		s.event = s.event.Uint16(key, i[0])
	}

	return s
}

func (s *event) Uint32(key string, i ...uint32) Event {
	if len(i) > 1 {
		s.event = s.event.Uints32(key, i)
	} else {
		s.event = s.event.Uint32(key, i[0])
	}

	return s
}

func (s *event) Uint64(key string, i ...uint64) Event {
	if len(i) > 1 {
		s.event = s.event.Uints64(key, i)
	} else {
		s.event = s.event.Uint64(key, i[0])
	}

	return s
}

func (s *event) Float32(key string, f ...float32) Event {
	if len(f) > 1 {
		s.event = s.event.Floats32(key, f)
	} else {
		s.event = s.event.Float32(key, f[0])
	}

	return s
}

func (s *event) Float64(key string, f ...float64) Event {
	if len(f) > 1 {
		s.event = s.event.Floats64(key, f)
	} else {
		s.event = s.event.Float64(key, f[0])
	}

	return s
}

func (s *event) Time(key string, t ...time.Time) Event {
	if len(t) > 1 {
		s.event = s.event.Times(key, t)
	} else {
		s.event = s.event.Time(key, t[0])
	}

	return s
}

func (s *event) Duration(key string, d ...time.Duration) Event {
	if len(d) > 1 {
		s.event = s.event.Durs(key, d)
	} else {
		s.event = s.event.Dur(key, d[0])
	}

	return s
}

func (s *event) Any(key string, i any) Event {
	s.event = s.event.Interface(key, i)
	return s
}

func (s *event) Enabled() bool {
	return s.event.Enabled()
}

func (s *event) Discard() Event {
	s.event = s.event.Discard()
	return s
}

func (s *event) Msg(msg string) {
	s.event.Msg(msg)
}

func (s *event) Send() {
	s.event.Send()
}

func (s *event) Msgf(format string, v ...interface{}) {
	s.event.Msgf(format, v...)
}

func (s *event) Bytes(key string, val []byte) Event {
	s.event = s.event.Bytes(key, val)
	return s
}

func (s *event) Hex(key string, val []byte) Event {
	s.event = s.event.Hex(key, val)
	return s
}

func (s *event) Stack() Event {
	s.event = s.event.Stack()
	return s
}

func (s *event) Timestamp() Event {
	s.event = s.event.Timestamp()
	return s
}

func (s *event) TimeDiff(key string, t time.Time, start time.Time) Event {
	s.event = s.event.TimeDiff(key, t, start)
	return s
}

func (s *event) Caller(skip ...int) Event {
	s.event = s.event.Caller(skip...)
	return s
}

func (s *event) IPAddr(key string, ip net.IP) Event {
	s.event = s.event.IPAddr(key, ip)
	return s
}

func (s *event) IPPrefix(key string, pfx net.IPNet) Event {
	s.event = s.event.IPPrefix(key, pfx)
	return s
}

func (s *event) MACAddr(key string, ha net.HardwareAddr) Event {
	s.event = s.event.MACAddr(key, ha)
	return s
}
