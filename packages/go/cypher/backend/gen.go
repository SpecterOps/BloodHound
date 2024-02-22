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

package backend

import (
	"bytes"
	"github.com/specterops/bloodhound/cypher/backend/cypher"
	"github.com/specterops/bloodhound/cypher/frontend"
	"github.com/specterops/bloodhound/cypher/model"
	"io"
)

type Emitter interface {
	Write(query *model.RegularQuery, writer io.Writer) error
	WriteExpression(output io.Writer, expression model.Expression) error
}

func CypherToCypher(ctx *frontend.Context, input string) (string, error) {
	if query, err := frontend.ParseCypher(ctx, input); err != nil {
		return "", err
	} else {
		var (
			output  = &bytes.Buffer{}
			emitter = cypher.Emitter{
				StripLiterals: false,
			}
		)

		if err := emitter.Write(query, output); err != nil {
			return "", err
		}

		return output.String(), nil
	}
}
