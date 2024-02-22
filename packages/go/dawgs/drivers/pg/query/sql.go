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

package query

import (
	"embed"
	"fmt"
	"path"
	"strings"
)

var (
	//go:embed sql
	queryFS embed.FS
)

func stripSQLComments(multiLineContent string) string {
	builder := strings.Builder{}

	for _, line := range strings.Split(multiLineContent, "\n") {
		trimmedLine := strings.TrimSpace(line)

		// Strip empty and SQL comment lines
		if len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "--") {
			continue
		}

		builder.WriteString(trimmedLine)
		builder.WriteString("\n")
	}

	return builder.String()
}

func readFile(name string) string {
	if content, err := queryFS.ReadFile(name); err != nil {
		panic(fmt.Sprintf("Unable to find embedded query file %s: %v", name, err))
	} else {
		return stripSQLComments(string(content))
	}
}

func loadSQL(name string) string {
	return readFile(path.Join("sql", name))
}

var (
	sqlSchemaUp           = loadSQL("schema_up.sql")
	sqlSchemaDown         = loadSQL("schema_down.sql")
	sqlSelectTableIndexes = loadSQL("select_table_indexes.sql")
	sqlSelectKindID       = loadSQL("select_table_indexes.sql")
	sqlSelectGraphs       = loadSQL("select_graphs.sql")
	sqlInsertGraph        = loadSQL("insert_graph.sql")
	sqlInsertKind         = loadSQL("insert_or_get_kind.sql")
	sqlSelectKinds        = loadSQL("select_kinds.sql")
	sqlSelectGraphByName  = loadSQL("select_graph_by_name.sql")
)
