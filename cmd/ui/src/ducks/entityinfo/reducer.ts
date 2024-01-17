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
import * as types from './types';

const initialEntityInfoState: types.EntityInfoState = {
    open: false,
    selectedNode: null,
};

const entityInfoReducer = (state = initialEntityInfoState, action: types.EntityInfoActionTypes) => {
    return produce(state, (draft) => {
        if (action.type === types.ENTITY_INFO_OPEN) {
            draft.open = action.open;
        } else if (action.type === types.SET_SELECTED_NODE) {
            draft.open = true;
            draft.selectedNode = action.selectedNode;
        }
    });
};

export default entityInfoReducer;
