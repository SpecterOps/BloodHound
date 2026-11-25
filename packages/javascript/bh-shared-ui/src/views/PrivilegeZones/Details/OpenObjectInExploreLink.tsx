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

import { TextButtonVariants } from 'doodle-ui';
import { AppLink } from '../../../components/Navigation/AppLink';
import { ExploreQueryParams, encodeCypherQuery } from '../../../hooks';
import { cn, createTypedSearchParams } from '../../../utils';

export const OpenObjectInExploreLink = ({ nodeId }: { nodeId: number }) => {
    if (!Number.isSafeInteger(nodeId)) return null;

    return (
        <AppLink
            to={{
                pathname: '/explore',
                search: createTypedSearchParams<ExploreQueryParams>({
                    cypherSearch: encodeCypherQuery(`MATCH (n) WHERE id(n) = ${nodeId} RETURN n`),
                    selectedItem: String(nodeId),
                    searchType: 'cypher',
                    exploreSearchTab: 'cypher',
                }),
            }}
            target='_blank'
            rel='noopener'
            aria-label='Open in Explore (opens in a new tab)'
            className={cn(TextButtonVariants({ fontColor: 'primary' }), '-ms-2 text-sm font-bold leading-7')}>
            Open in Explore
        </AppLink>
    );
};
