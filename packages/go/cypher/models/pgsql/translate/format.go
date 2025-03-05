// Copyright 2024 Specter Ops, Inc.
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

package translate

import (
	"bytes"
	"context"
	"github.com/specterops/bloodhound/cypher/models/cypher"
	cypherFormat "github.com/specterops/bloodhound/cypher/models/cypher/format"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/pgsql/format"
)

func Translated(translation Result) (string, error) {
	return format.Statement(translation.Statement, format.NewOutputBuilder())
}

func FromCypher(ctx context.Context, regularQuery *cypher.RegularQuery, kindMapper pgsql.KindMapper, stripLiterals bool) (format.Formatted, error) {
	var (
		output  = &bytes.Buffer{}
		emitter = cypherFormat.NewCypherEmitter(stripLiterals)
	)

	output.WriteString("-- ")

	if err := emitter.Write(regularQuery, output); err != nil {
		return format.Formatted{}, err
	}

	output.WriteString("\n")

	if translation, err := Translate(ctx, regularQuery, kindMapper, nil); err != nil {
		return format.Formatted{}, err
	} else if sqlQuery, err := format.Statement(translation.Statement, format.NewOutputBuilder()); err != nil {
		return format.Formatted{}, err
	} else {
		output.WriteString(sqlQuery)

		return format.Formatted{
			Statement:  output.String(),
			Parameters: translation.Parameters,
		}, nil
	}
}
