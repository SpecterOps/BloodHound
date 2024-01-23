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

package vendormocks

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../LICENSE.header -destination=./neo4j/neo4j-go-driver/v5/neo4j/mock.go -package=neo4j github.com/neo4j/neo4j-go-driver/v5/neo4j Result,Transaction,Session
//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../LICENSE.header -destination=./jackc/pgx/v5/mock.go -package=pgx github.com/jackc/pgx/v5 Tx
