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

import { produce } from 'immer';
import * as types from 'src/ducks/tierzero/types';

const initialTierZeroState: types.ChangeLogState = {
    tierZeroSelection: { id: null, type: null },
    changelog: {},
    error: false,
};

const tierZeroReducer = (state: types.ChangeLogState = initialTierZeroState, action: types.TierZeroActionTypes) => {
    return produce(state, (draft) => {
        switch (action.type) {
            case types.ADD_PRINCIPAL:
                draft.changelog[action.id] = {
                    id: action.id,
                    name: action.name,
                    change: 'add',
                    nodeType: action.nodeType,
                };
                break;
            case types.REMOVE_PRINCIPAL:
                draft.changelog[action.id] = {
                    id: action.id,
                    name: action.name,
                    change: 'remove',
                    nodeType: action.nodeType,
                };
                break;
            case types.UNDO_PRINCIPAL:
                delete draft.changelog[action.id];
                break;
            case types.ERROR:
                draft.error = true;
                break;
            case types.TIME_PASSED:
                break;
            case types.FLUSH_START:
                break;
            case types.FLUSH_SUCCESS:
                draft.changelog = {};
                draft.error = false;
                break;
            case types.DISCARD:
                draft.changelog = {};
                break;
            case types.CHECK:
                break;
            case types.SET_TIER_ZERO_SELECTION:
                draft.tierZeroSelection = { id: action.domainId, type: action.domainType };
                break;
            default:
                break;
        }
    });
};

export default tierZeroReducer;
