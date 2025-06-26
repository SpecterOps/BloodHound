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

module github.com/specterops/bloodhound/packages/go/stbernard

go 1.23.0

toolchain go1.23.8

require (
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/specterops/bloodhound/bhlog v0.0.0-00010101000000-000000000000
	github.com/specterops/bloodhound/slicesext v0.0.0-00010101000000-000000000000
	github.com/specterops/dawgs v0.1.3
	github.com/stretchr/testify v1.10.0
	golang.org/x/mod v0.24.0
)

require (
	github.com/RoaringBitmap/roaring/v2 v2.5.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/axiomhq/hyperloglog v0.2.5 // indirect
	github.com/bits-and-blooms/bitset v1.12.0 // indirect
	github.com/dgryski/go-metro v0.0.0-20180109044635-280f6062b5bc // indirect
	github.com/gammazero/deque v1.0.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgtype v1.14.4 // indirect
	github.com/jackc/pgx/v5 v5.7.5 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/kamstrup/intmap v0.5.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/neo4j/neo4j-go-driver/v5 v5.9.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/text v0.25.0 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/tools v0.32.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/specterops/bloodhound/bhlog => ../bhlog
	github.com/specterops/bloodhound/slicesext => ../slicesext
)
