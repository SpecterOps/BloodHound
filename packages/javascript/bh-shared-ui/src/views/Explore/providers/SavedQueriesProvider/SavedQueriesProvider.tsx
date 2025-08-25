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
import { createContext, Dispatch, SetStateAction, useContext, useState } from 'react';
import { useCypherSearch, useGetSelectedQuery } from '../../../../hooks';
import { QueryLineItem } from '../../../../types';
type SelectedType = {
    query: string;
    id?: number;
};
type SaveAction = 'edit' | 'save-as' | undefined;

interface SavedQueriesContextType {
    selected: SelectedType;
    selectedQuery: QueryLineItem | undefined;
    showSaveQueryDialog: boolean;
    saveAction: SaveAction;
    setSelected: Dispatch<SetStateAction<SelectedType>>;
    setShowSaveQueryDialog: Dispatch<SetStateAction<boolean>>;
    setSaveAction: Dispatch<SetStateAction<SaveAction>>;
    runQuery: any;
    editQuery: any;
}

export const SavedQueriesContext = createContext<SavedQueriesContextType>({
    selected: { query: '', id: undefined },
    selectedQuery: undefined,
    showSaveQueryDialog: false,
    saveAction: undefined,
    setSelected: (value) => {},
    setShowSaveQueryDialog: () => {},
    runQuery: (query: string, id: number) => {},
    editQuery: (id: number) => {},
    setSaveAction: (value) => {},
});

export const useSavedQueriesContext = () => {
    const context = useContext(SavedQueriesContext);
    if (!context) {
        throw new Error('MyContext provider is missing!');
    }
    return context;
};

export const SavedQueriesProvider = ({ children }: { children: any }) => {
    const [selected, setSelected] = useState<SelectedType>({ query: '', id: undefined });
    const [showSaveQueryDialog, setShowSaveQueryDialog] = useState(false);
    const [saveAction, setSaveAction] = useState<SaveAction>(undefined);

    const selectedQuery: QueryLineItem | undefined = useGetSelectedQuery(selected.query, selected.id);

    const { setCypherQuery, performSearch } = useCypherSearch();

    const runQuery = (query: string, id: number) => {
        setSelected({ query, id });
        setCypherQuery(query);
        performSearch(query);
    };

    const editQuery = (id: number) => {
        setSelected({ query: '', id: id });
        setSaveAction('edit');
        setShowSaveQueryDialog(true);
    };

    const contextValue = {
        selected,
        selectedQuery,
        saveAction,
        showSaveQueryDialog,
        setSelected,
        setSaveAction,
        setShowSaveQueryDialog,
        runQuery,
        editQuery,
    };

    return <SavedQueriesContext.Provider value={contextValue}>{children}</SavedQueriesContext.Provider>;
};
