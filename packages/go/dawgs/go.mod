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

module github.com/specterops/bloodhound/dawgs

go 1.23

require (
	github.com/RoaringBitmap/roaring v1.9.4
	github.com/axiomhq/hyperloglog v0.2.0
	github.com/gammazero/deque v0.2.1
	github.com/jackc/pgtype v1.14.4
	github.com/jackc/pgx/v5 v5.7.1
	github.com/neo4j/neo4j-go-driver/v5 v5.9.0
	github.com/specterops/bloodhound/bhlog v0.0.0-00010101000000-000000000000
	github.com/specterops/bloodhound/cypher v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.10.0
	go.uber.org/mock v0.5.1
)

require (
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/bits-and-blooms/bitset v1.12.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-metro v0.0.0-20211217172704-adc40b04c140 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	golang.org/x/crypto v0.35.0 // indirect
	golang.org/x/exp v0.0.0-20230515195305-f3d0a9c9a5cc // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/specterops/bloodhound/bhlog => ../bhlog
	github.com/specterops/bloodhound/cypher => ../cypher
)
