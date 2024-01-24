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

package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	iso8601 "github.com/channelmeter/iso8601duration"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/migrations"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
)

const (
	DefaultServerShutdownTimeout = time.Minute
	ContentSecurityPolicy        = "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self' data:;"
)

func NewDaemonContext(parentCtx context.Context) context.Context {
	daemonContext, doneFunc := context.WithCancel(parentCtx)

	go func() {
		defer doneFunc()

		// Shutdown on SIGINT/SIGTERM
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, syscall.SIGTERM)
		signal.Notify(signalChannel, syscall.SIGINT)

		// Wait for a signal from the OS
		<-signalChannel
	}()

	return daemonContext
}

// MigrateGraph runs migrations for the graph database
func MigrateGraph(ctx context.Context, db graph.Database, schema graph.Schema) error {
	return migrations.NewGraphMigrator(db).Migrate(ctx, schema)
}

// MigrateDB runs database migrations on PG
func MigrateDB(cfg config.Configuration, db database.Database) error {
	if err := db.Migrate(); err != nil {
		return err
	}

	if hasInstallation, err := db.HasInstallation(); err != nil {
		return err
	} else if hasInstallation {
		return nil
	}

	secretDigester := cfg.Crypto.Argon2.NewDigester()

	if roles, err := db.GetAllRoles("", model.SQLFilter{}); err != nil {
		return fmt.Errorf("error while attempting to fetch user roles: %w", err)
	} else if secretDigest, err := secretDigester.Digest(cfg.DefaultAdmin.Password); err != nil {
		return fmt.Errorf("error while attempting to digest secret for user: %w", err)
	} else if adminRole, found := roles.FindByName(auth.RoleAdministrator); !found {
		return fmt.Errorf("unable to find admin role")
	} else {
		var (
			adminUser = model.User{
				Roles: model.Roles{
					adminRole,
				},
				PrincipalName: cfg.DefaultAdmin.PrincipalName,
				EmailAddress:  null.NewString(cfg.DefaultAdmin.EmailAddress, true),
				FirstName:     null.NewString(cfg.DefaultAdmin.FirstName, true),
				LastName:      null.NewString(cfg.DefaultAdmin.LastName, true),
			}

			authSecret = model.AuthSecret{
				Digest:       secretDigest.String(),
				DigestMethod: secretDigester.Method(),
			}
		)

		if cfg.DefaultAdmin.ExpireNow {
			authSecret.ExpiresAt = time.Time{}
		} else if defaultWindow, err := iso8601.FromString(appcfg.DefaultPasswordExpirationWindow); err != nil {
			return fmt.Errorf("unable to parse default password expiration window: %w", err)
		} else {
			authSecret.ExpiresAt = time.Now().Add(defaultWindow.ToDuration())
		}

		if _, err := db.InitializeSecretAuth(adminUser, authSecret); err != nil {
			return fmt.Errorf("error in database while initalizing auth: %w", err)
		} else {
			passwordMsg := fmt.Sprintf("# Initial Password Set To:    %s    #", cfg.DefaultAdmin.Password)
			paddingString := strings.Repeat(" ", len(passwordMsg)-2)
			borderString := strings.Repeat("#", len(passwordMsg))

			log.Infof("%s", borderString)
			log.Infof("#%s#", paddingString)
			log.Infof("%s", passwordMsg)
			log.Infof("#%s#", paddingString)
			log.Infof("%s", borderString)
		}
	}

	return nil
}
