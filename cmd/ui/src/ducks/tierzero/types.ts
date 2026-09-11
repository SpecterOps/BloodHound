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

import { GraphNodeTypes } from '../graph/types';

export const ADD_PRINCIPAL = 'app/tierzero/ADD_PRINCIPAL';
export const REMOVE_PRINCIPAL = 'app/tierzero/REMOVE_PRINCIPAL';
export const UNDO_PRINCIPAL = 'app/tierzero/UNDO_PRINCIPAL';
export const ERROR = 'app/tierzero/ERROR';
export const TIME_PASSED = 'app/tierzero/TIME_PASSED';
export const FLUSH_START = 'app/tierzero/FLUSH_START';
export const FLUSH_SUCCESS = 'app/tierzero/FLUSH_SUCCESS';
export const DISCARD = 'app/tierzero/DISCARD';
export const CHECK = 'app/tierzero/CHECK';
export const SET_TIER_ZERO_SELECTION = 'app/explore/SET_TIER_ZERO_SELECTION';

export interface ChangeLogState {
    tierZeroSelection: { id: string | null; type: string | null };
    changelog: Record<string, { id: string; name: string; change: 'add' | 'remove'; nodeType: GraphNodeTypes }>;
    error: boolean;
}

export interface AddAction {
    type: typeof ADD_PRINCIPAL;
    id: string;
    name: string;
    nodeType: GraphNodeTypes;
}

export interface RemoveAction {
    type: typeof REMOVE_PRINCIPAL;
    id: string;
    name: string;
    nodeType: GraphNodeTypes;
}

export interface UndoAction {
    type: typeof UNDO_PRINCIPAL;
    id: string;
    nodeType: GraphNodeTypes;
}

export interface ErrorAction {
    type: typeof ERROR;
}

export interface FlushStartAction {
    type: typeof FLUSH_START;
}

export interface FlushSuccessAction {
    type: typeof FLUSH_SUCCESS;
}

export interface DiscardAction {
    type: typeof DISCARD;
}

export interface TimePassedAction {
    type: typeof TIME_PASSED;
    value: number;
}

export interface CheckAction {
    type: typeof CHECK;
    timePassed: number;
}

export interface SetTierZeroSelectionAction {
    type: typeof SET_TIER_ZERO_SELECTION;
    domainId: string | null;
    domainType: string | null;
}

export type TierZeroActionTypes =
    | ErrorAction
    | AddAction
    | RemoveAction
    | UndoAction
    | FlushStartAction
    | FlushSuccessAction
    | DiscardAction
    | TimePassedAction
    | CheckAction
    | SetTierZeroSelectionAction;
