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

package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
)

// With takes a function returning a mux.MiddlewareFunc type and applies it the to variadic list of routes
// passed in.
func With(limiterFactory func() mux.MiddlewareFunc, routes ...*Route) {
	for _, route := range routes {
		route.Use(limiterFactory())
	}
}

// Router is a wrapper for the mux.Router type. It adds service-specific functionality to HTTP handler routes created.
type Router struct {
	globalMiddleware []mux.MiddlewareFunc
	mux              *mux.Router
	authorizer       auth.Authorizer
}

// Route represents a route to a http.Handler. The handler is stored, wrapped by a middleware.Wrapper struct to allow
// for easier middleware registration that may be unique for each route.
type Route struct {
	handler    *middleware.Wrapper
	mux        *mux.Route
	authorizer auth.Authorizer
}

func (s *Route) Queries(pairs ...string) *Route {
	s.mux.Queries(pairs...)
	return s
}

func (s *Route) Methods(methods ...string) *Route {
	s.mux.Methods(methods...)
	return s
}

func (s *Route) Use(middleware ...mux.MiddlewareFunc) {
	s.handler.Use(middleware...)
}

func (s *Route) RequireAuth() *Route {
	return s.RequirePermissions()
}

// Ensure that the requestor has all of the listed permissions
func (s *Route) RequirePermissions(permissions ...model.Permission) *Route {
	s.handler.Use(middleware.PermissionsCheckAll(s.authorizer, permissions...))
	return s
}

// Ensure that the requestor has at least one of the listed permissions
func (s *Route) RequireAtLeastOnePermission(permissions ...model.Permission) *Route {
	s.handler.Use(middleware.PermissionsCheckAtLeastOne(s.authorizer, permissions...))
	return s
}

func (s *Route) AuthorizeUserManagementAccess() *Route {
	s.handler.Use(middleware.AuthorizeAuthManagementAccess(auth.Permissions(), s.authorizer))
	return s
}

func (s *Route) RequireUserId() *Route {
	s.handler.Use(middleware.RequireUserId())
	return s
}

// SupportsETAC wraps the ETAC middleware which allows or denies a user access to an environment (domainid, tenantid), when it is used in a route's path parameter
func (s *Route) SupportsETAC(db database.Database, dogTagsService dogtags.Service) *Route {
	s.handler.Use(middleware.SupportsETACMiddleware(db, dogTagsService))
	return s
}

func (s *Route) RequireAllEnvironmentAccess(dogTagsService dogtags.Service) *Route {
	s.handler.Use(middleware.RequireAllEnvironmentAccessMiddleware(dogTagsService))
	return s
}

func (s *Route) CheckFeatureFlag(db database.Database, flagKey string) *Route {
	s.handler.Use(middleware.FeatureFlagMiddleware(db, flagKey))
	return s
}

func NewRouter(cfg config.Configuration, authorizer auth.Authorizer, contentSecurityPolicy string) Router {
	muxRouter := mux.NewRouter()
	muxRouter.Use(middleware.EnsureRequestBodyClosed())
	muxRouter.Use(middleware.SecureHandlerMiddleware(cfg, contentSecurityPolicy))

	return Router{mux: muxRouter, authorizer: authorizer}
}

// UsePostrouting appends all of the given mux.MiddlewareFunc instances to this router's post-route middleware execution
// chain. Post-route means that this middleware will only be executed if a registered route is found to match the client
// request.
func (s Router) UsePostrouting(middleware ...mux.MiddlewareFunc) {
	s.mux.Use(middleware...)
}

// UsePrerouting appends all of the given mux.MiddlewareFunc instances to this router's pre-route middleware execution
// chain. Pre-route means that this middleware will only be executed for each request regardless of whether or not it
// matches to a valid route.
func (s *Router) UsePrerouting(middleware ...mux.MiddlewareFunc) {
	s.globalMiddleware = append(s.globalMiddleware, middleware...)
}

func (s Router) Handler() http.Handler {
	var handlerCursor http.Handler = s.mux

	// Wrap the cursor with the middleware in reverse
	for idx := len(s.globalMiddleware) - 1; idx >= 0; idx-- {
		handlerCursor = s.globalMiddleware[idx](handlerCursor)
	}

	return handlerCursor
}

func (s Router) PathPrefix(template string, handler http.Handler) *Route {
	middlewareWrapper := middleware.NewWrapper(handler)

	return &Route{
		handler: middlewareWrapper,
		mux:     s.mux.PathPrefix(template).Handler(middlewareWrapper),
	}
}

func (s Router) HandleFunc(template string, handlerFunc func(http.ResponseWriter, *http.Request)) *Route {
	middlewareWrapper := middleware.NewWrapper(http.HandlerFunc(handlerFunc))

	return &Route{
		handler:    middlewareWrapper,
		mux:        s.mux.Handle(template, middlewareWrapper),
		authorizer: s.authorizer,
	}
}

func (s Router) GET(template string, handlerFunc func(http.ResponseWriter, *http.Request)) *Route {
	return s.HandleFunc(template, handlerFunc).Methods(http.MethodGet)
}

func (s Router) PUT(template string, handlerFunc func(http.ResponseWriter, *http.Request)) *Route {
	return s.HandleFunc(template, handlerFunc).Methods(http.MethodPut)
}

func (s Router) POST(template string, handlerFunc func(http.ResponseWriter, *http.Request)) *Route {
	return s.HandleFunc(template, handlerFunc).Methods(http.MethodPost)
}

func (s Router) DELETE(template string, handlerFunc func(http.ResponseWriter, *http.Request)) *Route {
	return s.HandleFunc(template, handlerFunc).Methods(http.MethodDelete)
}

func (s Router) PATCH(template string, handlerFunc func(http.ResponseWriter, *http.Request)) *Route {
	return s.HandleFunc(template, handlerFunc).Methods(http.MethodPatch)
}
