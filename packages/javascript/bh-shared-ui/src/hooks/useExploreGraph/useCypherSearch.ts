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

import { useEffect, useState } from 'react';
import { useExploreParams } from '../useExploreParams';
import { decodeCypherQuery, encodeCypherQuery } from './utils';

export const useCypherSearch = () => {
    const [cypherQuery, setCypherQuery] = useState<string>('');

    const { cypherSearch, setExploreParams } = useExploreParams();

    useEffect(() => {
        if (cypherSearch) {
            const decoded = decodeCypherQuery(cypherSearch);
            setCypherQuery(decoded);
        } else {
            setCypherQuery('');
        }
    }, [cypherSearch]);

    // create query param with a query string if it is passed, and the field state otherwise
    const performSearch = (query?: string) => {
        setExploreParams({
            searchType: 'cypher',
            cypherSearch: encodeCypherQuery(query ?? cypherQuery),
        });
    };

    return {
        cypherQuery,
        setCypherQuery,
        performSearch,
    };
};
