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

import { takeEvery, all, call, put } from 'redux-saga/effects';
import * as actions from './actions';
import { apiClient } from 'bh-shared-ui';
import {
    GetAssetGroup,
    GetAssetGroupFailure,
    GetAssetGroupSuccess,
    ListAssetGroupsFailure,
    ListAssetGroupsSuccess,
} from './actionCreators';
import { GetAssetGroupAction } from './types';
import { SagaIterator } from 'redux-saga';

export function* listAssetGroupsSaga(): SagaIterator {
    try {
        const listAssetGroupsResponse = yield call(apiClient.listAssetGroups);
        const assetGroups = listAssetGroupsResponse.data.data.asset_groups;
        yield put(ListAssetGroupsSuccess(assetGroups));
        for (const assetGroup of assetGroups) {
            const assetGroupId = assetGroup.id;
            yield put(GetAssetGroup(assetGroupId));
        }
    } catch (error) {
        console.error(error);
        yield put(ListAssetGroupsFailure(error));
    }
}

export function* getAssetGroupSaga(payload: GetAssetGroupAction): SagaIterator {
    try {
        const getAssetGroupResponse = yield call(apiClient.getAssetGroup, payload.assetGroupId);
        yield put(GetAssetGroupSuccess(payload.assetGroupId, getAssetGroupResponse.data.data));
    } catch (error) {
        console.error(error);
        yield put(GetAssetGroupFailure(error));
    }
}

export default function* StartAssetGroupsSagas() {
    yield all([
        takeEvery(actions.LIST_ASSET_GROUPS, listAssetGroupsSaga),
        takeEvery(actions.GET_ASSET_GROUP, getAssetGroupSaga),
    ]);
}
