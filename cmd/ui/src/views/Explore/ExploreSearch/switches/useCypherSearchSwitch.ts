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

import { searchbarActions } from 'bh-shared-ui';
import { useAppDispatch, useAppSelector } from 'src/store';

export const useCypherSearchSwitch = () => {
    const dispatch = useAppDispatch();
    const reduxCypherQuery = useAppSelector((state) => state.search.cypher.searchTerm);

    return {
        cypherQuery: reduxCypherQuery,
        setCypherQuery: (query: string) => dispatch(searchbarActions.cypherQueryEdited(query)),
        performSearch: (query?: string) => dispatch(searchbarActions.cypherSearch(query ?? reduxCypherQuery)),
    };
};
