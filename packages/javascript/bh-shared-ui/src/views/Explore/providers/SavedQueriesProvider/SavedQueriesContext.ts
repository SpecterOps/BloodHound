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
import { createContext, Dispatch, SetStateAction, useContext } from 'react';
import { QueryLineItem, SaveQueryAction, SelectedQuery } from '../../../../types';

interface SavedQueriesContextType {
    selected: SelectedQuery;
    selectedQuery: QueryLineItem | undefined;
    showSaveQueryDialog: boolean;
    saveAction: SaveQueryAction | undefined;
    setSelected: Dispatch<SetStateAction<SelectedQuery>>;
    setShowSaveQueryDialog: Dispatch<SetStateAction<boolean>>;
    setSaveAction: Dispatch<SetStateAction<SaveQueryAction | undefined>>;
    runQuery: (query: string, id?: number) => void;
    editQuery: (id: number) => void;
}

export const SavedQueriesContext = createContext<SavedQueriesContextType>({
    selected: { query: '', id: undefined },
    selectedQuery: undefined,
    showSaveQueryDialog: false,
    saveAction: undefined,
    setSelected: () => {},
    setShowSaveQueryDialog: () => {},
    runQuery: () => {},
    editQuery: () => {},
    setSaveAction: () => {},
});

export const useSavedQueriesContext = () => {
    const context = useContext(SavedQueriesContext);
    if (!context) {
        throw new Error('MyContext provider is missing!');
    }
    return context;
};
