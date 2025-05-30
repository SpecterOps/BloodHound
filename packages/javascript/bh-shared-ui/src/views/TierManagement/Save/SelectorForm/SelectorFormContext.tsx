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

import { AssetGroupTagSelector, SeedTypeObjectId, SeedTypes, SelectorSeedRequest } from 'js-client-library';
import { createContext } from 'react';
import { UseQueryResult } from 'react-query';
import { Action, AssetGroupSelectedNodes } from './SelectorForm';

interface SelectorFormContext {
    dispatch: React.Dispatch<Action>;
    seeds: SelectorSeedRequest[];
    selectedObjects: AssetGroupSelectedNodes;
    selectorType: SeedTypes;
    selectorQuery: UseQueryResult<AssetGroupTagSelector>;
}

export const initialValue: SelectorFormContext = {
    dispatch: () => {},
    seeds: [],
    selectedObjects: [],
    selectorType: SeedTypeObjectId,
    selectorQuery: {
        data: undefined,
        isLoading: true,
        isError: false,
        isSuccess: false,
    } as UseQueryResult<AssetGroupTagSelector>,
};

const SelectorFormContext = createContext<SelectorFormContext>(initialValue);

export default SelectorFormContext;
