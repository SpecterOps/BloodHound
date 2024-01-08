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

package test

import (
	"context"
	"time"
)

type Controller interface {
	Cleanup(func())
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	FailNow()
}

type Context interface {
	Controller
	context.Context
}

type controllerInstance struct {
	Controller
	ctx context.Context
}

func (s controllerInstance) Deadline() (deadline time.Time, ok bool) {
	return s.ctx.Deadline()
}

func (s controllerInstance) Done() <-chan struct{} {
	return s.ctx.Done()
}

func (s controllerInstance) Err() error {
	return s.ctx.Err()
}

func (s controllerInstance) Value(key any) any {
	return s.ctx.Value(key)
}

func (s controllerInstance) Context() context.Context {
	return s
}

func WithContext(parentCtx context.Context, controller Controller) Context {
	testCtx, doneFunc := context.WithCancel(parentCtx)

	// Ensure the done function for the context is called by test cleanup
	controller.Cleanup(doneFunc)

	return controllerInstance{
		Controller: controller,
		ctx:        testCtx,
	}
}

func NewContext(controller Controller) Context {
	return WithContext(context.Background(), controller)
}
