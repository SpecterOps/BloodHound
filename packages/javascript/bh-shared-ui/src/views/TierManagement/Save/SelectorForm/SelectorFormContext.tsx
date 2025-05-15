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

import {
    AssetGroupTagNode,
    AssetGroupTagSelector,
    SeedTypeObjectId,
    SeedTypes,
    SelectorSeedRequest,
} from 'js-client-library';
import { createContext } from 'react';
import { UseQueryResult } from 'react-query';

interface SelectorFormContext {
    seeds: SelectorSeedRequest[];
    setSeeds: React.Dispatch<React.SetStateAction<SelectorSeedRequest[]>>;
    results: AssetGroupTagNode[] | null;
    setResults: React.Dispatch<React.SetStateAction<AssetGroupTagNode[] | null>>;
    selectorType: SeedTypes;
    setSelectorType: React.Dispatch<React.SetStateAction<SeedTypes>>;
    selectorQuery: UseQueryResult<AssetGroupTagSelector>;
}

export const initialValue: SelectorFormContext = {
    seeds: [],
    setSeeds: () => {},
    results: null,
    setResults: () => {},
    selectorType: SeedTypeObjectId,
    setSelectorType: () => {},
    selectorQuery: {
        data: undefined,
        isLoading: true,
        isError: false,
        isSuccess: false,
    } as UseQueryResult<AssetGroupTagSelector>,
};

const SelectorFormContext = createContext<SelectorFormContext>(initialValue);

export default SelectorFormContext;
