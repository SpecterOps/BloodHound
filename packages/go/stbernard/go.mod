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
	github.com/specterops/bloodhound/dawgs v0.0.0-00010101000000-000000000000
	github.com/specterops/bloodhound/graphschema v0.0.0-00010101000000-000000000000
	github.com/specterops/bloodhound/slicesext v0.0.0-00010101000000-000000000000
	github.com/specterops/bloodhound/src v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.10.0
	golang.org/x/mod v0.21.0
)

require (
	github.com/RoaringBitmap/roaring v1.9.4 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/axiomhq/hyperloglog v0.2.0 // indirect
	github.com/beevik/etree v1.2.0 // indirect
	github.com/bits-and-blooms/bitset v1.12.0 // indirect
	github.com/bloodhoundad/azurehound/v2 v2.4.1 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/channelmeter/iso8601duration v0.0.0-20150204201828-8da3af7a2a61 // indirect
	github.com/crewjam/httperr v0.2.0 // indirect
	github.com/crewjam/saml v0.4.14 // indirect
	github.com/dgryski/go-metro v0.0.0-20211217172704-adc40b04c140 // indirect
	github.com/gammazero/deque v0.2.1 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gobeam/stringy v0.0.6 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgtype v1.14.4 // indirect
	github.com/jackc/pgx/v5 v5.7.1 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20230326075908-cb1d2100619a // indirect
	github.com/mattermost/xml-roundtrip-validator v0.1.0 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/neo4j/neo4j-go-driver/v5 v5.9.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/power-devops/perfstat v0.0.0-20221212215047-62379fc7944b // indirect
	github.com/pquerna/otp v1.4.0 // indirect
	github.com/russellhaering/goxmldsig v1.4.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.1 // indirect
	github.com/shirou/gopsutil/v3 v3.23.5 // indirect
	github.com/specterops/bloodhound/analysis v0.0.0-00010101000000-000000000000 // indirect
	github.com/specterops/bloodhound/bomenc v0.0.0-00010101000000-000000000000 // indirect
	github.com/specterops/bloodhound/cache v0.0.0-00010101000000-000000000000 // indirect
	github.com/specterops/bloodhound/crypto v0.0.0-00010101000000-000000000000 // indirect
	github.com/specterops/bloodhound/cypher v0.0.0-00010101000000-000000000000 // indirect
	github.com/specterops/bloodhound/ein v0.0.0-00010101000000-000000000000 // indirect
	github.com/specterops/bloodhound/headers v0.0.0-00010101000000-000000000000 // indirect
	github.com/specterops/bloodhound/lab v0.0.0-00010101000000-000000000000 // indirect
	github.com/specterops/bloodhound/mediatypes v0.0.0-00010101000000-000000000000 // indirect
	github.com/specterops/bloodhound/openapi v0.0.0-00010101000000-000000000000 // indirect
	github.com/specterops/bloodhound/params v0.0.0-00010101000000-000000000000 // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	golang.org/x/crypto v0.35.0 // indirect
	golang.org/x/exp v0.0.0-20230515195305-f3d0a9c9a5cc // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	gorm.io/gorm v1.25.12 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/tools v0.26.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/specterops/bloodhound/analysis => ../analysis
	github.com/specterops/bloodhound/bhlog => ../bhlog
	github.com/specterops/bloodhound/bomenc => ../bomenc
	github.com/specterops/bloodhound/cache => ../cache
	github.com/specterops/bloodhound/crypto => ../crypto
	github.com/specterops/bloodhound/cypher => ../cypher
	github.com/specterops/bloodhound/dawgs => ../dawgs
	github.com/specterops/bloodhound/ein => ../ein
	github.com/specterops/bloodhound/graphschema => ../graphschema
	github.com/specterops/bloodhound/headers => ../headers
	github.com/specterops/bloodhound/lab => ../lab
	github.com/specterops/bloodhound/mediatypes => ../mediatypes
	github.com/specterops/bloodhound/openapi => ../openapi
	github.com/specterops/bloodhound/params => ../params
	github.com/specterops/bloodhound/slicesext => ../slicesext
	github.com/specterops/bloodhound/src => ../../../cmd/api/src

)
