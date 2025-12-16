// Copyright 2025 Specter Ops, Inc.

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
import { createContext, useContext } from 'react';
import { DetailsTabOption, TagOption } from './utils';

export interface SelectedDetailsTabContext {
    selectedDetailsTab: DetailsTabOption;
    setSelectedDetailsTab: (tabValue: DetailsTabOption) => void;
}

const initialSelectedDetailsTabValue = {
    selectedDetailsTab: TagOption,
    setSelectedDetailsTab: () => {},
};

export const SelectedDetailsTabContext = createContext<SelectedDetailsTabContext>(initialSelectedDetailsTabValue);

export const useSelectedDetailsTabContext = () => {
    const context = useContext(SelectedDetailsTabContext);

    if (!context) {
        throw new Error('useHistoryTableContext is outside of SelectedDetailsTabContext');
    }

    return context;
};
