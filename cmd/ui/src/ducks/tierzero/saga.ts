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

import find from 'lodash/find';
import keys from 'lodash/keys';
import { SagaIterator } from 'redux-saga';
import { all, call, put, select, takeEvery } from 'redux-saga/effects';
import { apiClient } from 'bh-shared-ui';
import { addSnackbar } from 'src/ducks/global/actions';
import { flushSuccess } from 'src/ducks/tierzero/actions';
import { FLUSH_START } from 'src/ducks/tierzero/types';
import type { AppState } from 'src/store';

function* flushWorker(): SagaIterator {
    try {
        const assetGroups: any[] = yield select((state: AppState) => state.assetgroups.assetGroups);

        const tierZeroAssetGroupId = find(assetGroups, (assetGroup) => {
            return assetGroup.tag === 'admin_tier_0';
        })?.id;

        if (tierZeroAssetGroupId === undefined) throw new Error('Tier Zero Asset Group ID could not be identified');

        const changelog: Record<
            string,
            {
                id: string;
                name: string;
                change: 'add' | 'remove';
            }
        > = yield select((state: AppState) => state.tierzero.changelog);

        const selectors = keys(changelog).map((objectid) => ({
            selector_name: objectid,
            sid: objectid,
            action: changelog[objectid].change,
        }));

        yield call(apiClient.updateAssetGroupSelector, tierZeroAssetGroupId, selectors);

        yield put(
            addSnackbar(
                'Update successful. Please check back later to view updated Asset Group.',
                'tierZeroFlushSuccess'
            )
        );

        yield put(flushSuccess());
    } catch (error) {
        console.error(error);
    }
}

export default function* startTierZeroSagas() {
    yield all([takeEvery(FLUSH_START, flushWorker)]);
}
