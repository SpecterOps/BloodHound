// Copyright 2026 Specter Ops, Inc.
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

// Package safetemplate defines the restricted set of utility functions
// available to author-supplied Go templates, such as KindInfo markdown
// templates uploaded as part of an OpenGraph extension. It is the single
// source of truth shared by validation paths and rendering paths.
package safetemplate

import (
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// we don't support these template functions
var unsupportedFns = []string{
	"bcrypt",
	"htpasswd",
	"genPrivateKey",
	"derivePassword",
	"buildCustomCert",
	"genCA",
	"genCAWithKey",
	"genSelfSignedCert",
	"genSignedCert",
	"genSignedCertWithKey",
	"encryptAES",
	"decryptAES",
	"regexMatch",
	"regexFindAll",
	"regexFind",
	"regexReplaceAll",
	"regexReplaceAllLiteral",
	"regexSplit",
	"mustRegexMatch",
	"mustRegexFindAll",
	"mustRegexFind",
	"mustRegexReplaceAll",
	"mustRegexReplaceAllLiteral",
	"mustRegexSplit",
	"urlParse",
	"urlJoin",
	"randInt",
	"ago",
}

// FuncMap returns the utility functions available to author-supplied
// templates: the sprig hermetic text function map minus the unsupported
// functions.
func FuncMap() template.FuncMap {
	functions := sprig.HermeticTxtFuncMap()

	for _, functionName := range unsupportedFns {
		delete(functions, functionName)
	}

	return functions
}
