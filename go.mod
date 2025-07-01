// Copyright 2025 Specter Ops, Inc.
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
module github.com/specterops/bloodhound

go 1.24.4

require (
	cuelang.org/go v0.13.2
	github.com/Masterminds/semver/v3 v3.3.1
	github.com/RoaringBitmap/roaring v1.9.4
	github.com/bloodhoundad/azurehound/v2 v2.6.0
	github.com/channelmeter/iso8601duration v0.0.0-20150204201828-8da3af7a2a61
	github.com/coreos/go-oidc/v3 v3.14.1
	github.com/crewjam/saml v0.5.1
	github.com/dave/jennifer v1.7.1
	github.com/go-chi/chi/v5 v5.2.2
	github.com/gobeam/stringy v0.0.7
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/golang-jwt/jwt/v4 v4.5.2
	github.com/google/go-cmp v0.7.0
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/schema v1.4.1
	github.com/hashicorp/golang-lru v1.0.2
	github.com/jackc/pgx/v5 v5.7.5
	github.com/jedib0t/go-pretty/v6 v6.6.7
	github.com/neo4j/neo4j-go-driver/v5 v5.28.1
	github.com/peterldowns/pgtestdb v0.1.1
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.5.0
	github.com/prometheus/client_golang v1.22.0
	github.com/russellhaering/goxmldsig v1.5.0
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/specterops/dawgs v0.1.5
	github.com/stretchr/testify v1.10.0
	github.com/teambition/rrule-go v1.8.2
	github.com/ulule/limiter/v3 v3.11.2
	github.com/unrolled/secure v1.17.0
	go.uber.org/mock v0.5.2
	golang.org/x/crypto v0.39.0
	golang.org/x/mod v0.25.0
	golang.org/x/oauth2 v0.30.0
	golang.org/x/text v0.26.0
	golang.org/x/tools v0.34.0
	gorm.io/driver/postgres v1.5.10
	gorm.io/gorm v1.25.12
)

require (
	cuelabs.dev/go/oci/ociregistry v0.0.0-20250304105642-27e071d2c9b1 // indirect
	github.com/RoaringBitmap/roaring/v2 v2.5.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/axiomhq/hyperloglog v0.2.5 // indirect
	github.com/beevik/etree v1.5.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.12.0 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cockroachdb/apd/v3 v3.2.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-metro v0.0.0-20180109044635-280f6062b5bc // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/emicklei/proto v1.14.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/gammazero/deque v1.0.0 // indirect
	github.com/go-jose/go-jose/v4 v4.0.5 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgtype v1.14.4 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/kamstrup/intmap v0.5.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mattermost/xml-roundtrip-validator v0.1.0 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/protocolbuffers/txtpbfmt v0.0.0-20250129171521-feedd8250727 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	golang.org/x/exp v0.0.0-20240909161429-701f63a606c0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

tool github.com/specterops/bloodhound/packages/go/stbernard
