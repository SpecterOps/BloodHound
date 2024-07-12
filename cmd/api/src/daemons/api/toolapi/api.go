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

package toolapi

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/api/tools"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/bootstrap"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"net/http"
	"net/http/pprof"
	"strconv"
	"time"
)

// Daemon holds data relevant to the tools API daemon
type Daemon struct {
	cfg    config.Configuration
	server *http.Server
}

const supportUserEmail = "support@specterops.io"

func assertAndGetSupportUser(ctx context.Context, db database.Database) (model.User, error) {
	if supportUser, err := db.LookupUser(ctx, supportUserEmail); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			if roles, err := db.GetAllRoles(ctx, "", model.SQLFilter{}); err != nil {
				return model.User{}, err
			} else {
				var (
					adminRole model.Role
					found     = false
				)

				for _, role := range roles {
					if role.Name == auth.RoleAdministrator {
						adminRole = role
						found = true
					}
				}

				if !found {
					return model.User{}, fmt.Errorf("can't find admin role")
				}

				return db.CreateUser(ctx, model.User{
					Roles: []model.Role{
						adminRole,
					},
					FirstName:     null.NewString("SpecterOps", true),
					LastName:      null.NewString("Support", true),
					EmailAddress:  null.NewString(supportUserEmail, true),
					PrincipalName: "support",
					IsDisabled:    false,
					EULAAccepted:  true,
				})
			}
		} else {
			return model.User{}, err
		}
	} else {
		return supportUser, nil
	}
}

func invalidateSupportUserSession(ctx context.Context, db database.Database, sessionID int64) error {
	if activeSession, err := db.GetUserSession(ctx, sessionID); err != nil {
		return err
	} else if activeSession.User.EmailAddress.String != supportUserEmail {
		return fmt.Errorf("invalid user - only support user account sessions may be invalidated")
	} else {
		db.EndUserSession(ctx, activeSession)
	}

	return nil
}

func invalidateAllSupportUserSessions(ctx context.Context, db database.Database) error {
	if user, err := assertAndGetSupportUser(ctx, db); err != nil {
		return err
	} else if activeUserSessions, err := db.LookupActiveSessionsByUser(ctx, user); err != nil {
		return err
	} else {
		for _, activeUserSession := range activeUserSessions {
			db.EndUserSession(ctx, activeUserSession)
		}
	}

	return nil
}

type UserSession struct {
	SignedJWT string
	Session   model.UserSession
}

func createSupportUserSession(ctx context.Context, cfg config.Configuration, db database.Database) (UserSession, error) {
	if user, err := assertAndGetSupportUser(ctx, db); err != nil {
		return UserSession{}, err
	} else {
		userSession := model.UserSession{
			User:             user,
			UserID:           user.ID,
			ExpiresAt:        time.Now().UTC().Add(cfg.AuthSessionTTL()),
			AuthProviderType: model.SessionAuthProviderJIT,
			AuthProviderID:   0,
		}

		if newSession, err := db.CreateUserSession(ctx, userSession); err != nil {
			return UserSession{}, err
		} else if signingKeyBytes, err := cfg.Crypto.JWT.SigningKeyBytes(); err != nil {
			return UserSession{}, err
		} else {
			var (
				jwtClaims = &auth.SessionData{
					StandardClaims: jwt.StandardClaims{
						Id:        strconv.FormatInt(newSession.ID, 10),
						Subject:   user.ID.String(),
						IssuedAt:  newSession.CreatedAt.UTC().Unix(),
						ExpiresAt: newSession.ExpiresAt.UTC().Unix(),
					},
				}

				token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
			)

			if signedToken, err := token.SignedString(signingKeyBytes); err != nil {
				return UserSession{}, err
			} else {
				return UserSession{
					SignedJWT: signedToken,
					Session:   newSession,
				}, nil
			}
		}
	}
}

func Audit(ctx context.Context, db database.Database, action model.AuditLogAction, status model.AuditLogEntryStatus, data model.AuditData) error {
	if newEntry, err := model.NewAuditEntry(action, status, data); err != nil {
		return err
	} else {
		return db.AppendAuditLog(ctx, newEntry)
	}
}

func AuditIntent(ctx context.Context, db database.Database, action model.AuditLogAction, data model.AuditData) error {
	return Audit(ctx, db, action, model.AuditLogStatusIntent, data)
}

func AuditSuccess(ctx context.Context, db database.Database, action model.AuditLogAction, data model.AuditData) error {
	return Audit(ctx, db, action, model.AuditLogStatusSuccess, data)
}

func AuditFailure(ctx context.Context, db database.Database, action model.AuditLogAction, data model.AuditData) error {
	return Audit(ctx, db, action, model.AuditLogStatusFailure, data)
}

func UnderAudit(ctx context.Context, db database.Database, action model.AuditLogAction, data model.AuditData, logic func() error) error {
	if err := AuditIntent(ctx, db, action, data); err != nil {
		return err
	}

	if err := logic(); err != nil {
		if err := AuditFailure(ctx, db, action, data); err != nil {
			log.Errorf("Unable to commit audit failure write: %v", err)
		}

		return err
	}

	if err := AuditSuccess(ctx, db, action, data); err != nil {
		log.Errorf("Unable to commit audit success write: %v", err)
	}

	return nil
}

func NewDaemon[DBType database.Database](ctx context.Context, connections bootstrap.DatabaseConnections[DBType, *graph.DatabaseSwitch], cfg config.Configuration, graphSchema graph.Schema, extensions ...func(router *chi.Mux)) Daemon {
	var (
		pgMigrator    = tools.NewPGMigrator(ctx, cfg, graphSchema, connections.Graph)
		router        = chi.NewRouter()
		toolContainer = tools.NewToolContainer(connections.RDMS)
	)

	router.Mount("/metrics", promhttp.Handler())

	// Support normal pprof endpoints for easier consumption with standard tools
	router.Mount("/debug/pprof/allocs", pprof.Handler("allocs"))
	router.Mount("/debug/pprof/block", pprof.Handler("block"))
	router.Mount("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Mount("/debug/pprof/heap", pprof.Handler("heap"))
	router.Mount("/debug/pprof/mutex", pprof.Handler("mutex"))
	router.Mount("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Get("/debug/pprof/profile", pprof.Profile)
	router.Get("/debug/pprof/trace", pprof.Trace)

	// TODO: remove old trace handler when we can wire up acumen to handle the above pprof endpoints instead
	router.Get("/trace", tools.NewTraceHandler())

	router.Put("/graph-db/switch/pg", pgMigrator.SwitchPostgreSQL)
	router.Put("/graph-db/switch/neo4j", pgMigrator.SwitchNeo4j)
	router.Put("/pg-migration/start", pgMigrator.MigrationStart)
	router.Get("/pg-migration/status", pgMigrator.MigrationStatus)
	router.Put("/pg-migration/cancel", pgMigrator.MigrationCancel)

	router.Get("/logging", tools.GetLoggingDetails)
	router.Put("/logging", tools.PutLoggingDetails)

	if cfg.EnableJITSupportAccess {
		router.Post("/auth/support", func(response http.ResponseWriter, request *http.Request) {
			if err := UnderAudit(request.Context(), connections.RDMS, "CreateSupportLoginSession", map[string]any{}, func() error {
				if supportUserSession, err := createSupportUserSession(ctx, cfg, connections.RDMS); err != nil {
					return err
				} else {
					api.WriteJSONResponse(request.Context(), map[string]any{
						"session_id": supportUserSession.Session.ID,
						"token":      supportUserSession.SignedJWT,
					}, http.StatusCreated, response)
				}

				return nil
			}); err != nil {
				api.WriteJSONResponse(request.Context(), map[string]any{
					"error": err.Error(),
				}, http.StatusInternalServerError, response)
			}
		})

		router.Delete("/auth/support", func(response http.ResponseWriter, request *http.Request) {
			if err := UnderAudit(request.Context(), connections.RDMS, "InvalidateSupportLoginSessions", map[string]any{}, func() error {
				if err := invalidateAllSupportUserSessions(ctx, connections.RDMS); err != nil {
					return err
				}

				return nil
			}); err != nil {
				api.WriteJSONResponse(request.Context(), map[string]any{
					"error": err.Error(),
				}, http.StatusInternalServerError, response)
			} else {
				response.WriteHeader(http.StatusOK)
			}
		})

		router.Delete("/auth/support/{session_id:[0-9]+}", func(response http.ResponseWriter, request *http.Request) {
			sessionIDStr := chi.URLParam(request, "session_id")

			if sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64); err != nil {
				api.WriteJSONResponse(request.Context(), map[string]any{
					"error": err.Error(),
				}, http.StatusInternalServerError, response)
			} else if err := UnderAudit(request.Context(), connections.RDMS, "InvalidateSupportLoginSessions", map[string]any{}, func() error {
				if err := invalidateSupportUserSession(ctx, connections.RDMS, sessionID); err != nil {
					return err
				}

				return nil
			}); err != nil {
				api.WriteJSONResponse(request.Context(), map[string]any{
					"error": err.Error(),
				}, http.StatusInternalServerError, response)
			} else {
				response.WriteHeader(http.StatusOK)
			}
		})
	}

	router.Get("/features", toolContainer.GetFlags)
	router.Put("/features/{feature_id:[0-9]+}/toggle", toolContainer.ToggleFlag)

	for _, extension := range extensions {
		extension(router)
	}

	return Daemon{
		cfg: cfg,
		server: &http.Server{
			Addr:     cfg.MetricsPort,
			Handler:  router,
			ErrorLog: log.Adapter(log.LevelError, "ToolAPI", 0),
		},
	}
}

// Name returns the name of the daemon
func (s Daemon) Name() string {
	return "Tools API"
}

// Start begins the daemon and waits for a stop signal in the exit channel
func (s Daemon) Start(ctx context.Context) {
	if s.cfg.TLS.Enabled() {
		if err := s.server.ListenAndServeTLS(s.cfg.TLS.CertFile, s.cfg.TLS.KeyFile); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Errorf("HTTP server listen error: %v", err)
			}
		}
	} else {
		if err := s.server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Errorf("HTTP server listen error: %v", err)
			}
		}
	}
}

// Stop passes in a stop signal to the exit channel, thereby killing the daemon
func (s Daemon) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
