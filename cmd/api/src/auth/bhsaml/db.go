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

package bhsaml

import (
	"context"

	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/serde"
)

func formatSAMLProviderURLs(requestContext context.Context, samlProviders ...model.SAMLProvider) model.SAMLProviders {
	for idx := 0; idx < len(samlProviders); idx++ {
		providerURLs := FormatServiceProviderURLs(*ctx.Get(requestContext).Host, samlProviders[idx].Name)

		samlProviders[idx].ServiceProviderIssuerURI = serde.FromURL(providerURLs.ServiceProviderRoot)
		samlProviders[idx].ServiceProviderInitiationURI = serde.FromURL(providerURLs.SingleSignOnService)
		samlProviders[idx].ServiceProviderMetadataURI = serde.FromURL(providerURLs.MetadataService)
		samlProviders[idx].ServiceProviderACSURI = serde.FromURL(providerURLs.AssertionConsumerService)
	}

	return samlProviders
}

func GetSAMLProviderByName(db database.Database, name string, requestContext context.Context) (model.SAMLProvider, error) {
	if samlProvider, err := db.LookupSAMLProviderByName(name); err != nil {
		return model.SAMLProvider{}, err
	} else {
		return formatSAMLProviderURLs(requestContext, samlProvider)[0], nil
	}
}

func GetSAMLProviderByID(db database.Database, id int32, requestContext context.Context) (model.SAMLProvider, error) {
	if samlProvider, err := db.GetSAMLProvider(id); err != nil {
		return model.SAMLProvider{}, err
	} else {
		return formatSAMLProviderURLs(requestContext, samlProvider)[0], nil
	}
}

func GetAllSAMLProviders(db database.Database, requestContext context.Context) (model.SAMLProviders, error) {
	if samlProviders, err := db.GetAllSAMLProviders(); err != nil {
		return nil, err
	} else {
		return formatSAMLProviderURLs(requestContext, samlProviders...), nil
	}
}
