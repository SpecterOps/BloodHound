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

package validation

import (
	"strings"
)

type ValidatorFactory struct {
	factories map[string]ValidatorFactoryFunc
}

type ValidatorFactoryFunc func(options map[string]string) Validator

func (s ValidatorFactory) NewValidatorFromTag(tag string) Validator {
	params := strings.Split(tag, ",")
	validatorType := params[0]

	options := map[string]string{}
	for _, kv := range params[1:] {
		tuple := strings.Split(kv, "=")
		options[tuple[0]] = tuple[1]
	}

	if validatorFactory, ok := s.factories[validatorType]; !ok {
		return nil
	} else {
		return validatorFactory(options)
	}
}

var validatorFactory ValidatorFactory

func init() {
	validatorFactory = ValidatorFactory{map[string]ValidatorFactoryFunc{
		"password": NewPasswordValidator,
		"required": NewRequiredValidator,
	}}
}
