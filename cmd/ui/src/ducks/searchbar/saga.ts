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

import { SagaIterator } from 'redux-saga';
import { all, call, fork, put, select, takeLatest } from 'redux-saga/effects';
import apiClient from 'src/api';
import { startCypherQuery, startPathfindingQuery, startSearchQuery } from 'src/ducks/explore/actions';
import * as actions from 'src/ducks/searchbar/actions';
import * as types from 'src/ducks/searchbar/types';
import type { AppState } from 'src/store';

function* searchWatcher(): SagaIterator {
    yield takeLatest(types.SEARCH_START, searchWorker);
}

function* startSearchActionWatcher(): SagaIterator {
    yield takeLatest(types.SEARCH_SELECTED, searchSelectedWorker);
}

function* searchWorker(payload: types.SearchStartAction) {
    if (payload.target === types.CYPHER_SEARCH) {
        yield put(startCypherQuery(payload.searchTerm));
    } else {
        try {
            const { data } = yield call(apiClient.searchHandler, payload.searchTerm);
            yield put(actions.searchSuccessAction(data as types.SearchNodeType[], payload.target));
        } catch {
            /* empty */
        }
    }
}

function* searchSelectedWorker(payload: types.StartSearchSelectedAction) {
    const searchState: types.SearchState = yield select((state: AppState) => state.search);
    const edges = searchState.pathFilters
        .filter((pathFilter) => pathFilter.checked)
        .map((pathFilter) => pathFilter.edgeType);

    if (
        payload.target === types.SECONDARY_SEARCH &&
        searchState.primary.value !== null &&
        searchState.secondary.value !== null
    ) {
        yield put(
            startPathfindingQuery(searchState.primary.value.objectid, searchState.secondary.value.objectid, edges)
        );
    } else if (payload.target === types.PRIMARY_SEARCH && searchState.primary.value !== null) {
        yield put(startSearchQuery(searchState.primary.value.objectid, searchState.searchType));
    }
}

export default function* startSearchSagas() {
    yield all([fork(searchWatcher), fork(startSearchActionWatcher)]);
}
