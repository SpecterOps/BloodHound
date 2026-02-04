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
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
)

const (
	DefaultServerShutdownTimeout = time.Minute
	ContentSecurityPolicy        = "default-src 'self'; script-src 'self' %s 'unsafe-inline'; style-src 'self' %s 'unsafe-inline'; img-src 'self' %s data: blob:; connect-src 'self' %s; frame-src 'self' %s; font-src 'self' %s data:;"
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

// MigrateDB runs database migrations on PG
func MigrateDB(ctx context.Context, cfg config.Configuration, db database.Database, defaultAdminFunc func() (config.DefaultAdminConfiguration, error)) error {
	if err := db.Migrate(ctx); err != nil {
		return err
	}

	if hasInstallation, err := db.HasInstallation(ctx); err != nil {
		return err
	} else if hasInstallation {
		return nil
	}

	return CreateDefaultAdmin(ctx, cfg, db, defaultAdminFunc)
}

func PopulateExtensionData(ctx context.Context, db database.Database) error {
	if err := db.PopulateExtensionData(ctx); err != nil {
		return err
	}

	return nil
}

// FillAndPopulateDefaultAdminInfo will ensure that the default admin config has all of the necessary values for population in the DB
func FillAndPopulateDefaultAdminInfo(cfg config.DefaultAdminConfiguration, defaultAdminFunction func() (config.DefaultAdminConfiguration, error)) (config.DefaultAdminConfiguration, bool, error) {
	if cfg.PrincipalName == "" || cfg.Password == "" || cfg.EmailAddress == "" || cfg.FirstName == "" || cfg.LastName == "" {
		if defaultAdminConfiguration, err := defaultAdminFunction(); err != nil {
			return cfg, false, fmt.Errorf("error in setup initializing auth secret: %w", err)
		} else {
			var needsLog = false
			if cfg.PrincipalName == "" {
				cfg.PrincipalName = defaultAdminConfiguration.PrincipalName
			}

			if cfg.Password == "" {
				cfg.Password = defaultAdminConfiguration.Password
				needsLog = true
			}

			if cfg.EmailAddress == "" {
				cfg.EmailAddress = defaultAdminConfiguration.EmailAddress
			}

			if cfg.FirstName == "" {
				cfg.FirstName = defaultAdminConfiguration.FirstName
			}

			if cfg.LastName == "" {
				cfg.LastName = defaultAdminConfiguration.LastName
			}

			return cfg, needsLog, nil
		}
	} else {
		return cfg, false, nil
	}
}

func CreateDefaultAdmin(ctx context.Context, cfg config.Configuration, db database.Database, defaultAdminFunction func() (config.DefaultAdminConfiguration, error)) error {
	var (
		secretDigester = cfg.Crypto.Argon2.NewDigester()
		needsLog       bool
	)

	//Populate any missing fields of the admin configuration using our defaults
	if populatedConfig, needsLogInner, err := FillAndPopulateDefaultAdminInfo(cfg.DefaultAdmin, defaultAdminFunction); err != nil {
		return fmt.Errorf("error while populating default admin info: %w", err)
	} else {
		cfg.DefaultAdmin = populatedConfig
		needsLog = needsLogInner
	}

	if roles, err := db.GetAllRoles(ctx, "", model.SQLFilter{}); err != nil {
		return fmt.Errorf("error while attempting to fetch user roles: %w", err)
	} else if secretDigest, err := secretDigester.Digest(cfg.DefaultAdmin.Password); err != nil {
		return fmt.Errorf("error while attempting to digest secret for user: %w", err)
	} else if adminRole, found := roles.FindByName(auth.RoleAdministrator); !found {
		return fmt.Errorf("unable to find admin role")
	} else {
		if existingUser, err := db.LookupUser(ctx, cfg.DefaultAdmin.PrincipalName); !errors.Is(err, database.ErrNotFound) && err != nil {
			return fmt.Errorf("unable to lookup existing admin user: %w", err)
		} else if err == nil {
			if err := db.DeleteUser(ctx, existingUser); err != nil {
				return fmt.Errorf("unable to delete exisiting admin user: %s: %w", existingUser.PrincipalName, err)
			}
		}

		var (
			adminUser = model.User{
				Roles: model.Roles{
					adminRole,
				},
				PrincipalName:   cfg.DefaultAdmin.PrincipalName,
				EmailAddress:    null.NewString(cfg.DefaultAdmin.EmailAddress, true),
				FirstName:       null.NewString(cfg.DefaultAdmin.FirstName, true),
				LastName:        null.NewString(cfg.DefaultAdmin.LastName, true),
				AllEnvironments: true,
			}

			authSecret = model.AuthSecret{
				Digest:       secretDigest.String(),
				DigestMethod: secretDigester.Method(),
			}
		)

		if cfg.DefaultAdmin.ExpireNow {
			authSecret.ExpiresAt = time.Time{}
		} else {
			authSecret.ExpiresAt = time.Now().Add(appcfg.GetPasswordExpiration(ctx, db))
		}

		if _, err := db.InitializeSecretAuth(ctx, adminUser, authSecret); err != nil {
			return fmt.Errorf("error in database while initializing auth: %w", err)
		} else if needsLog {
			passwordMsg := fmt.Sprintf("# Initial Password Set To:    %s    #", cfg.DefaultAdmin.Password)
			paddingString := strings.Repeat(" ", len(passwordMsg)-2)
			borderString := strings.Repeat("#", len(passwordMsg))

			fmt.Println(borderString)
			fmt.Printf("#%s#\n", paddingString)
			fmt.Println(passwordMsg)
			fmt.Printf("#%s#\n", paddingString)
			fmt.Println(borderString)
		} else {
			fmt.Printf("Password has been set from existing config or environment variable\n")
		}
	}

	return nil
}
