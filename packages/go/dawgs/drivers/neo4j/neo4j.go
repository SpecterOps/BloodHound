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

package neo4j

import (
	"fmt"
	"net/url"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/channels"
)

const (
	defaultNeo4jTransactionTimeout = time.Minute * 15
)

func newNeo4jDB(connectionURLStr string) (graph.Database, error) {
	if connectionURL, err := url.Parse(connectionURLStr); err != nil {
		return nil, err
	} else if connectionURL.Scheme != DriverName {
		return nil, fmt.Errorf("expected connection URL scheme %s for Neo4J but got %s", DriverName, connectionURL.Scheme)
	} else if password, isSet := connectionURL.User.Password(); !isSet {
		return nil, fmt.Errorf("no password provided in connection URL")
	} else {
		boltURL := fmt.Sprintf("bolt://%s:%s", connectionURL.Hostname(), connectionURL.Port())

		if internalDriver, err := neo4j.NewDriver(boltURL, neo4j.BasicAuth(connectionURL.User.Username(), password, "")); err != nil {
			return nil, fmt.Errorf("unable to connect to Neo4J: %w", err)
		} else {
			return &driver{
				driver:                    internalDriver,
				defaultTransactionTimeout: defaultNeo4jTransactionTimeout,
				limiter:                   channels.NewConcurrencyLimiter(DefaultConcurrentConnections),
				writeFlushSize:            DefaultWriteFlushSize,
				batchWriteSize:            DefaultBatchWriteSize,
			}, nil
		}
	}
}

func init() {
	dawgs.Register(DriverName, func(cfg any) (graph.Database, error) {
		if connectionURLStr, typeOK := cfg.(string); !typeOK {
			return nil, fmt.Errorf("expected string for configuration type but got %T", cfg)
		} else {
			return newNeo4jDB(connectionURLStr)
		}
	})
}
