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

package channels

import (
	"context"
)

// Submit takes a valid context.Context, a writeable channel and attempts to submit the passed value to it. If the
// function successfully submits the value before the context expires, this function will return true. If, however the
// context expires before the function can submit to the channel this function return false. Context expiration is not
// explicitly considered an error. It is up to the caller to interpret the return value further.
func Submit[T any](ctx context.Context, channel chan<- T, value T) bool {
	select {
	case channel <- value:
		return true

	case <-ctx.Done():
		return false
	}
}

// Receive takes a valid context.Context, a readable channel and attempts to receive a new value from it. If the
// function successfully receives the value before the context expires, this function will return both the value and
// true. If, however the context expires before the function can receive from the channel, or if the channel is closed,
// this function return false. Both conditions are not explicitly considered errors. It is up to the caller to interpret
// the return values further.
func Receive[T any](ctx context.Context, inC <-chan T) (T, bool) {
	var defaultValue T

	select {
	case value, hasNextValue := <-inC:
		if hasNextValue {
			return value, true
		}

	case <-ctx.Done():
	}

	return defaultValue, false
}

// Pipe is a context aware fused Receive and Submit operation. The function will attempt to receive from the passed
// readable channel and then submit it to the writable channel. If the context expires during either of these
// operations this function will return false.
func Pipe[T any](ctx context.Context, inC <-chan T, outC chan<- T) bool {
	if value, hasValue := Receive(ctx, inC); hasValue {
		return Submit(ctx, outC, value)
	}

	return false
}

// PipeAll is a convenience function that exhausts the passed readable channel by piping the received values to passed
// writable channel.
func PipeAll[T any](ctx context.Context, inC <-chan T, outC chan<- T) {
	for Pipe(ctx, inC, outC) {
	}
}

// Pipeline is a function that operates similarly to a Pipe but allows conversion from the readable channel's type: R
// to the writable channel's type: T through a passed convertor delegate. Errors from the convertor delegate are
// returned to the caller.
//
// Conversion failure will result in a receive from the readable channel without a corresponding submit to the writable
// channel.
func Pipeline[R any, T any](ctx context.Context, inC <-chan R, outC chan<- T, convertor func(raw R) (T, error)) (bool, error) {
	if value, hasValue := Receive(ctx, inC); hasValue {
		if convertedValue, err := convertor(value); err != nil {
			return false, err
		} else {
			return Submit(ctx, outC, convertedValue), nil
		}

	}

	return false, nil
}

// PipelineAll is a convenience function that exhausts the passed readable channel by pipelining the received values to
// the passed writable channel.
func PipelineAll[R any, T any](ctx context.Context, inC <-chan R, outC chan<- T, convertor func(raw R) (T, error)) error {
	for {
		if success, err := Pipeline(ctx, inC, outC, convertor); !success || err != nil {
			return err
		}
	}
}
