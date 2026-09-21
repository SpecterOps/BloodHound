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
package dogtags

import "errors"

// ErrProviderNotImplemented is returned by the no-op provider
var ErrProviderNotImplemented = errors.New("no-op provider not implemented")

// Provider is the interface for dogtag backends.
// Providers are simple key-value stores - they don't know about defaults.
// The service layer handles defaults when a key isn't found.
type Provider interface {
	Name() string
	GetFlagAsBool(key string) (bool, error)
	GetFlagAsString(key string) (string, error)
	GetFlagAsInt(key string) (int64, error)
}

// NoopProvider returns ErrNotFound for all keys.
// The service will use defaults for everything.
type NoopProvider struct{}

const NoopProviderName = "NoopProvider"

func NewNoopProvider() *NoopProvider {
	return &NoopProvider{}
}

func (p *NoopProvider) Name() string {
	return NoopProviderName
}

func (p *NoopProvider) GetFlagAsBool(key string) (bool, error) {
	return false, ErrProviderNotImplemented
}

func (p *NoopProvider) GetFlagAsString(key string) (string, error) {
	return "", ErrProviderNotImplemented
}

func (p *NoopProvider) GetFlagAsInt(key string) (int64, error) {
	return 0, ErrProviderNotImplemented
}
