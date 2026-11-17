// Copyright 2026 Specter Ops, Inc.
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

package appdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/auth/internal/services"
)

const (
	TableSSOProviders  = "sso_providers"
	TableOIDCProviders = "oidc_providers"
	TableSAMLProviders = "saml_providers"

	ssoProviderColumnID         = "id"
	ssoProviderColumnName       = "name"
	ssoProviderColumnType       = "type"
	ssoProviderColumnSlug       = "slug"
	ssoProviderColumnConfig     = "config"
	ssoProviderColumnCreatedAt  = "created_at"
	ssoProviderColumnUpdatedAt  = "updated_at"
	providerColumnSSOProviderID = "sso_provider_id"

	ssoProviderAlias  = "sso"
	oidcProviderAlias = "oidc"
	samlProviderAlias = "saml"
)

var ssoProviderColumns = []string{
	ssoProviderColumnID,
	ssoProviderColumnName,
	ssoProviderColumnType,
	ssoProviderColumnSlug,
	ssoProviderColumnConfig,
}

type authProviderRow struct {
	ID           int32  `db:"id"`
	Name         string `db:"name"`
	Type         int    `db:"type"`
	Slug         string `db:"slug"`
	Config       []byte `db:"config"`
	OIDCProvider []byte `db:"oidc_provider"`
	SAMLProvider []byte `db:"saml_provider"`
}

func buildSSOProviderSort(sortItems params.SortItems, columns map[string]string) ([]string, error) {
	if len(sortItems) == 0 {
		return []string{qualify(ssoProviderAlias, ssoProviderColumnCreatedAt)}, nil
	}

	var orderByCols = make([]string, 0, len(sortItems))

	for _, item := range sortItems {
		if column, isSortable := columns[item.Field]; !isSortable {
			return nil, fmt.Errorf("%w: SSO provider sort references unknown field %q", services.ErrInvalidProviderQuery, item.Field)
		} else {
			orderByCols = append(orderByCols, column+" "+strings.ToUpper(string(item.Direction)))
		}
	}

	return orderByCols, nil
}

func buildSSOProviderFilter(queryFilters params.Filters, columns map[string]string) (*sqlbuilder.WhereClause, error) {
	var (
		condition   = sqlbuilder.NewCond()
		whereClause = sqlbuilder.NewWhereClause()
		fields      = make([]string, 0, len(queryFilters))
	)

	for field := range queryFilters {
		fields = append(fields, field)
	}

	sort.Strings(fields)

	for _, field := range fields {
		var (
			fieldFilters = queryFilters[field]
			andClauses   = make([]string, 0, len(fieldFilters))
			orClauses    = make([]string, 0, len(fieldFilters))
		)

		for _, filter := range fieldFilters {
			var clause string

			column, isFilterable := columns[filter.Field]
			if !isFilterable {
				return nil, fmt.Errorf("%w: SSO provider filter references unknown field %q", services.ErrInvalidProviderQuery, filter.Field)
			}

			// The middleware validates the field and operator but passes the value through as a raw string.
			// Convert the value to the column's type so that a bad value returns a 400 rather than a Postgres cast error surfaced as a 500
			var value any = filter.Value

			switch filter.Field {
			case "id", "type":
				if parsed, err := strconv.ParseInt(filter.Value, 10, 32); err != nil {
					return nil, fmt.Errorf("%w: %s filter value %q is not an integer", services.ErrInvalidProviderQuery, filter.Field, filter.Value)
				} else {
					value = int32(parsed)
				}
			case "created_at", "updated_at":
				if parsed, err := time.Parse(time.RFC3339Nano, filter.Value); err != nil {
					return nil, fmt.Errorf("%w: %s filter value %q is not an RFC3339 timestamp", services.ErrInvalidProviderQuery, filter.Field, filter.Value)
				} else {
					value = parsed
				}
			}

			switch filter.Operator {
			case params.Equals:
				clause = condition.Equal(column, value)
			case params.NotEquals:
				clause = condition.NotEqual(column, value)
			case params.GreaterThan:
				clause = condition.GreaterThan(column, value)
			case params.GreaterThanOrEquals:
				clause = condition.GreaterEqualThan(column, value)
			case params.LessThan:
				clause = condition.LessThan(column, value)
			case params.LessThanOrEquals:
				clause = condition.LessEqualThan(column, value)
			case params.ApproximatelyEquals:
				clause = condition.ILike(column, "%"+filter.Value+"%")
			default:
				return nil, fmt.Errorf("%w: SSO provider filter uses unsupported operator %q", services.ErrInvalidProviderQuery, filter.Operator)
			}

			if filter.SetOperator == params.FilterOr {
				orClauses = append(orClauses, clause)
			} else {
				andClauses = append(andClauses, clause)
			}
		}

		if len(andClauses) > 0 {
			whereClause.AddWhereExpr(condition.Args, andClauses...)
		}

		if len(orClauses) > 0 {
			whereClause.AddWhereExpr(condition.Args, condition.Or(orClauses...))
		}
	}

	return whereClause, nil
}

// ListAuthProviders loads both provider kinds and their protocol details
func (s *Store) ListAuthProviders(ctx context.Context, filters params.Filters, sortItems params.SortItems) ([]services.SSOProvider, error) {
	var (
		columns           = qualifyAll(ssoProviderAlias, ssoProviderColumns...)
		filterSortColumns = map[string]string{
			ssoProviderColumnID:        qualify(ssoProviderAlias, ssoProviderColumnID),
			ssoProviderColumnName:      qualify(ssoProviderAlias, ssoProviderColumnName),
			ssoProviderColumnSlug:      qualify(ssoProviderAlias, ssoProviderColumnSlug),
			ssoProviderColumnType:      qualify(ssoProviderAlias, ssoProviderColumnType),
			ssoProviderColumnCreatedAt: qualify(ssoProviderAlias, ssoProviderColumnCreatedAt),
			ssoProviderColumnUpdatedAt: qualify(ssoProviderAlias, ssoProviderColumnUpdatedAt),
		}
	)

	filterWhereClause, err := buildSSOProviderFilter(filters, filterSortColumns)
	if err != nil {
		return nil, err
	}

	sortOrderByCols, err := buildSSOProviderSort(sortItems, filterSortColumns)
	if err != nil {
		return nil, err
	}

	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	selectBuilder.Select(append(columns,
		selectBuilder.As("to_jsonb("+oidcProviderAlias+")", "oidc_provider"),
		selectBuilder.As("to_jsonb("+samlProviderAlias+")", "saml_provider"),
	)...)
	selectBuilder.From(selectBuilder.As(TableSSOProviders, ssoProviderAlias))
	selectBuilder.JoinWithOption(sqlbuilder.LeftJoin,
		selectBuilder.As(TableOIDCProviders, oidcProviderAlias),
		qualify(oidcProviderAlias, providerColumnSSOProviderID)+" = "+qualify(ssoProviderAlias, ssoProviderColumnID),
	)
	selectBuilder.JoinWithOption(sqlbuilder.LeftJoin,
		selectBuilder.As(TableSAMLProviders, samlProviderAlias),
		qualify(samlProviderAlias, providerColumnSSOProviderID)+" = "+qualify(ssoProviderAlias, ssoProviderColumnID),
	)
	// Retain OIDC visibility controlled by the legacy feature flag
	selectBuilder.Where("(sso.type = 1 OR (SELECT enabled FROM feature_flags WHERE key = 'oidc_support'))")
	selectBuilder.AddWhereClause(filterWhereClause)
	selectBuilder.OrderBy(sortOrderByCols...)

	selectQuery, args := selectBuilder.Build()

	if rows, err := s.db.Query(ctx, selectQuery, args...); err != nil {
		return nil, fmt.Errorf("listing SSO providers: %w", err)
	} else if providerRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[authProviderRow]); err != nil {
		return nil, fmt.Errorf("reading SSO providers: %w", err)
	} else {
		return toAuthProviders(providerRows)
	}
}

func toAuthProviders(rows []authProviderRow) ([]services.SSOProvider, error) {
	// Preserve the legacy nil slice when no providers are configured
	var providers []services.SSOProvider

	for _, row := range rows {
		if provider, err := toAuthProvider(row); err != nil {
			return nil, err
		} else {
			providers = append(providers, provider)
		}
	}

	return providers, nil
}

func toAuthProvider(row authProviderRow) (services.SSOProvider, error) {
	var (
		provider = services.SSOProvider{ID: row.ID, Name: row.Name, Type: services.ProviderType(row.Type), Slug: row.Slug}
		oidc     *OIDCProvider
		saml     *SAMLProvider
		config   ProviderConfig
	)

	if len(row.OIDCProvider) > 0 {
		if err := json.Unmarshal(row.OIDCProvider, &oidc); err != nil {
			return services.SSOProvider{}, fmt.Errorf("decoding OIDC provider: %w", err)
		} else if oidc != nil {
			provider.OIDCProvider = &services.OIDCProvider{
				ClientID:           oidc.ClientID,
				Issuer:             oidc.Issuer,
				SSOProviderID:      oidc.SSOProviderID,
				ProviderTimestamps: domainTimestamps(oidc.ProviderTimestamps),
			}
		}
	}

	if len(row.SAMLProvider) > 0 {
		if err := json.Unmarshal(row.SAMLProvider, &saml); err != nil {
			return services.SSOProvider{}, fmt.Errorf("decoding SAML provider: %w", err)
		} else if saml != nil {
			provider.SAMLProvider = &services.SAMLProvider{
				Name:                       saml.Name,
				DisplayName:                saml.DisplayName,
				IssuerURI:                  saml.IssuerURI,
				SingleSignOnURI:            saml.SingleSignOnURI,
				RootURIVersion:             saml.RootURIVersion,
				PrincipalAttributeMappings: saml.PrincipalAttributeMappings,
				SSOProviderID:              saml.SSOProviderID,
				ProviderTimestamps:         domainTimestamps(saml.ProviderTimestamps),
			}
		}
	}

	if len(row.Config) > 0 {
		if err := json.Unmarshal(row.Config, &config); err != nil {
			return services.SSOProvider{}, fmt.Errorf("decoding SSO config: %w", err)
		}
	}

	provider.Config = services.ProviderConfig{AutoProvision: services.AutoProvisionConfig(config.AutoProvision)}

	return provider, nil
}

func domainTimestamps(input ProviderTimestamps) services.ProviderTimestamps {
	var deletedAt sql.NullTime

	if input.DeletedAt != nil {
		deletedAt = sql.NullTime{Time: *input.DeletedAt, Valid: true}
	}

	return services.ProviderTimestamps{ID: input.ID, CreatedAt: input.CreatedAt, UpdatedAt: input.UpdatedAt, DeletedAt: deletedAt}
}
