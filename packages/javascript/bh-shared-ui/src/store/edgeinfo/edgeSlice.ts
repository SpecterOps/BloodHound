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

import { PayloadAction, createSlice } from '@reduxjs/toolkit';

export type EdgeInfoState = {
    open: boolean;
    selectedEdge: SelectedEdge;
    expandedSections: ExpandedEdgeSections;
};

export type SelectedEdge = {
    id: string;
    name: string;
    data: Record<string, any>;
    sourceNode: { name: string; id: string | number; objectId: string; type: string; haslaps?: boolean };
    targetNode: { name: string; id: string | number; objectId: string; type: string; haslaps?: boolean };
} | null;

export type ExpandedEdgeSections = Record<keyof typeof EdgeSections, boolean>;

export const EdgeSections = {
    data: 'Relationship Information',
    general: 'General',
    abuse: 'Abuse',
    windowsAbuse: 'Windows Abuse',
    linuxAbuse: 'Linux Abuse',
    opsec: 'OPSEC',
    references: 'References',
    composition: 'Composition',
} as const;

export const initialState: EdgeInfoState = {
    open: false,
    selectedEdge: null,
    expandedSections: {
        data: true,
        general: false,
        abuse: false,
        windowsAbuse: false,
        linuxAbuse: false,
        opsec: false,
        references: false,
        composition: false,
    },
};

export const edgeInfoSlice = createSlice({
    name: 'edgeinfo',
    initialState,
    reducers: {
        setEdgeInfoOpen: (state, action: PayloadAction<boolean>) => {
            state.open = action.payload;
        },
        setSelectedEdge: (state, action: PayloadAction<SelectedEdge>) => {
            state.selectedEdge = action.payload;
            Object.entries(state.expandedSections).forEach((section) => {
                const sectionKey = section[0] as keyof ExpandedEdgeSections;
                if (section[0] === 'data') state.expandedSections[sectionKey] = true;
                else state.expandedSections[sectionKey] = false;
            });
        },
        edgeSectionToggle: (
            state,
            action: PayloadAction<{ section: keyof typeof EdgeSections; expanded: boolean }>
        ) => {
            state.open = true;
            state.expandedSections[action.payload.section] = action.payload.expanded;
        },
        collapseAllSections: (state) => {
            Object.entries(state.expandedSections).forEach((section) => {
                const sectionKey = section[0] as keyof ExpandedEdgeSections;
                state.expandedSections[sectionKey] = false;
            });
        },
    },
});

export const { setEdgeInfoOpen, setSelectedEdge, edgeSectionToggle, collapseAllSections } = edgeInfoSlice.actions;

export default edgeInfoSlice.reducer;
