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

import { GraphNodeTypes } from 'src/ducks/graph/types';
import * as types from 'src/ducks/tierzero/types';

export const error = (): types.TierZeroActionTypes => {
    return {
        type: types.ERROR,
    };
};

export const addPrincipal = (id: string, name: string, nodeType: GraphNodeTypes): types.AddAction => {
    return {
        type: types.ADD_PRINCIPAL,
        id,
        name,
        nodeType,
    };
};

export const removePrincipal = (id: string, name: string, nodeType: GraphNodeTypes): types.RemoveAction => {
    return {
        type: types.REMOVE_PRINCIPAL,
        id,
        name,
        nodeType,
    };
};
export const undoPrincipal = (id: string, nodeType: GraphNodeTypes): types.UndoAction => {
    return {
        type: types.UNDO_PRINCIPAL,
        id,
        nodeType,
    };
};

export const flushStart = (): types.FlushStartAction => {
    return {
        type: types.FLUSH_START,
    };
};

export const flushSuccess = (): types.FlushSuccessAction => {
    return {
        type: types.FLUSH_SUCCESS,
    };
};

export const discardChanges = (): types.DiscardAction => {
    return {
        type: types.DISCARD,
    };
};

export const setTierZeroSelection = (
    domainId: string | null,
    domainType: string | null
): types.SetTierZeroSelectionAction => {
    return {
        type: types.SET_TIER_ZERO_SELECTION,
        domainId,
        domainType,
    };
};
