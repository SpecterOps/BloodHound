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

import {
    CYPHER_SEARCH,
    CypherSearchAction,
    DESTINATION_NODE_SELECTED,
    PATHFINDING_SEARCH,
    PRIMARY_SEARCH,
    SEARCH_TYPE_EXACT,
    SOURCE_NODE_SELECTED,
    SearchState,
    SourceNodeSelectedAction,
} from 'bh-shared-ui';
import { SagaIterator } from 'redux-saga';
import { all, fork, put, select, takeLatest } from 'redux-saga/effects';
import { startCypherQuery, startPathfindingQuery, startSearchQuery } from 'src/ducks/explore/actions';
import type { AppState } from 'src/store';

function* cypherSearchWatcher(): SagaIterator {
    yield takeLatest(CYPHER_SEARCH, cypherSearchWorker);
}

function* cypherSearchWorker(payload: CypherSearchAction) {
    if (payload.searchTerm) {
        yield put(startCypherQuery(payload.searchTerm));
    } else {
        const { cypher }: SearchState = yield select((state: AppState) => state.search);
        const { searchTerm } = cypher;
        if (searchTerm) {
            yield put(startCypherQuery(searchTerm));
        }
    }
}

function* primarySearchWatcher(): SagaIterator {
    yield takeLatest(SOURCE_NODE_SELECTED, primarySearchWorker);
    yield takeLatest(PRIMARY_SEARCH, primarySearchWorker);
}

function* primarySearchWorker(payload: SourceNodeSelectedAction) {
    const { primary, secondary, pathFilters }: SearchState = yield select((state: AppState) => state.search);

    const edges = pathFilters.filter((pathFilter) => pathFilter.checked).map((pathFilter) => pathFilter.edgeType);

    // try a pathfinding search first if flag is true
    if (payload.doPathfindSearch) {
        if (primary.value && secondary.value) {
            yield put(startPathfindingQuery(primary.value.objectid, secondary.value.objectid, edges));
        } else if (primary.value) {
            yield put(startSearchQuery(primary.value.objectid, SEARCH_TYPE_EXACT));
        }
    } else if (primary.value) {
        yield put(startSearchQuery(primary.value.objectid, SEARCH_TYPE_EXACT));
    }
}

function* pathfindingSearchWatcher(): SagaIterator {
    yield takeLatest(DESTINATION_NODE_SELECTED, pathfindingSearchWorker);
    yield takeLatest(PATHFINDING_SEARCH, pathfindingSearchWorker);
}

function* pathfindingSearchWorker() {
    const { primary, secondary, pathFilters }: SearchState = yield select((state: AppState) => state.search);

    const edges = pathFilters.filter((pathFilter) => pathFilter.checked).map((pathFilter) => pathFilter.edgeType);

    // first try a pathfinding search
    if (primary.value && secondary.value) {
        yield put(startPathfindingQuery(primary.value.objectid, secondary.value.objectid, edges));
    } else if (secondary.value) {
        // then try a primary search on the `destination` node
        yield put(startSearchQuery(secondary.value.objectid, SEARCH_TYPE_EXACT));
    } else if (primary.value) {
        // then try a primary search on the `source` node
        yield put(startSearchQuery(primary.value.objectid, SEARCH_TYPE_EXACT));
    }
}

export default function* startSearchSagas() {
    yield all([fork(primarySearchWatcher), fork(pathfindingSearchWatcher), fork(cypherSearchWatcher)]);
}
