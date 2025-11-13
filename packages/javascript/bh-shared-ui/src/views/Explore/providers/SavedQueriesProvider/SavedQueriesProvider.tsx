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
import { useState } from 'react';
import { useCypherSearch, useGetSelectedQuery } from '../../../../hooks';
import { QueryLineItem, SaveQueryAction } from '../../../../types';
import { SavedQueriesContext } from './SavedQueriesContext';

export function SavedQueriesProvider({ children }: { children: any }) {
    const { setCypherQuery, performSearch } = useCypherSearch();

    const [selectedId, setSelectedId] = useState<number | undefined>(undefined);
    const [showSaveQueryDialog, setShowSaveQueryDialog] = useState(false);
    const [saveAction, setSaveAction] = useState<SaveQueryAction | undefined>(undefined);

    const selectedQuery: QueryLineItem | undefined = useGetSelectedQuery(selectedId);

    const runQuery = (query: string, id?: number) => {
        setSelectedId(id);
        setCypherQuery(query);
        performSearch(query);
    };

    const editQuery = (id: number) => {
        setSelectedId(id);
        setSaveAction('edit');
        setShowSaveQueryDialog(true);
    };

    const contextValue = {
        selectedId,
        selectedQuery,
        saveAction,
        showSaveQueryDialog,
        setSelectedId,
        setSaveAction,
        setShowSaveQueryDialog,
        runQuery,
        editQuery,
    };

    return <SavedQueriesContext.Provider value={contextValue}>{children}</SavedQueriesContext.Provider>;
}
