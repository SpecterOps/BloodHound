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

package model

import (
	"github.com/specterops/bloodhound/dawgs/graph"
	"strconv"
	"strings"
)

const (
	NodeTable = "node"
	EdgeTable = "edge"
)

func partitionTableName(parent string, graphID int32) string {
	return parent + "_" + strconv.FormatInt(int64(graphID), 10)
}

func NodePartitionTableName(graphID int32) string {
	return partitionTableName(NodeTable, graphID)
}

func EdgePartitionTableName(graphID int32) string {
	return partitionTableName(EdgeTable, graphID)
}

func IndexName(table string, index graph.Index) string {
	stringBuilder := strings.Builder{}

	stringBuilder.WriteString(table)
	stringBuilder.WriteString("_")
	stringBuilder.WriteString(index.Field)
	stringBuilder.WriteString("_index")

	return stringBuilder.String()
}

func ConstraintName(table string, constraint graph.Constraint) string {
	stringBuilder := strings.Builder{}

	stringBuilder.WriteString(table)
	stringBuilder.WriteString("_")
	stringBuilder.WriteString(constraint.Field)
	stringBuilder.WriteString("_constraint")

	return stringBuilder.String()
}
