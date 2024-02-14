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

package fixtures

import (
	"fmt"
	"log"

	"github.com/specterops/bloodhound/src/auth"

	"github.com/specterops/bloodhound/lab"
	"github.com/specterops/bloodhound/src/bootstrap"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/test/integration"
)

var PostgresFixture = lab.NewFixture(func(harness *lab.Harness) (*database.BloodhoundDB, error) {
	if config, ok := lab.Unpack(harness, ConfigFixture); !ok {
		return nil, fmt.Errorf("unable to unpack ConfigFixture")
	} else if pgdb, err := database.OpenDatabase(config.Database.PostgreSQLConnectionString()); err != nil {
		return nil, err
	} else if err := integration.Prepare(database.NewBloodhoundDB(pgdb, auth.NewIdentityResolver())); err != nil {
		return nil, fmt.Errorf("failed ensuring database: %v", err)
	} else if err := bootstrap.MigrateDB(config, database.NewBloodhoundDB(pgdb, auth.NewIdentityResolver())); err != nil {
		return nil, fmt.Errorf("failed migrating database: %v", err)
	} else {
		return database.NewBloodhoundDB(pgdb, auth.NewIdentityResolver()), nil
	}
}, nil)

func begin() *lab.Fixture[*database.BloodhoundDB] {
	fixture := lab.NewFixture(func(harness *lab.Harness) (*database.BloodhoundDB, error) {
		if db, ok := lab.Unpack(harness, PostgresFixture); !ok {
			return nil, fmt.Errorf("unable to unpack BloodhoundDB")
		} else {
			return db.Begin(), nil
		}
	}, nil)

	if err := lab.SetDependency(fixture, PostgresFixture); err != nil {
		log.Fatal(err)
	}
	return fixture
}

func end(tx *lab.Fixture[*database.BloodhoundDB], dependencies []lab.Depender) {
	fixture := lab.NewFixture(func(harness *lab.Harness) (*database.BloodhoundDB, error) {
		if db, ok := lab.Unpack(harness, tx); !ok {
			return nil, fmt.Errorf("unable to unpack BloodhoundDB transaction")
		} else {
			return db, db.Commit()
		}
	}, nil)

	if err := lab.SetDependency(fixture, tx); err != nil {
		log.Fatal(err)
	} else {
		for _, dep := range dependencies {
			if err := lab.SetDependency(fixture, dep); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func TransactionalFixtures(factories ...func(*lab.Fixture[*database.BloodhoundDB]) lab.Depender) {
	tx := begin()
	var dependencies []lab.Depender
	for _, fn := range factories {
		dependencies = append(dependencies, fn(tx))
	}
	end(tx, dependencies)
}

func init() {
	if err := lab.SetDependency(PostgresFixture, ConfigFixture); err != nil {
		log.Fatalln(err)
	}
}
