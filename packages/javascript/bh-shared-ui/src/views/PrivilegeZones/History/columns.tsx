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

import { Tooltip, createColumnHelper } from '@bloodhoundenterprise/doodleui';
import { DateTime } from 'luxon';
import { LuxonFormat } from '../../..';
import { NoteCell } from './NoteCell';
import { HistoryItem } from './types';

const columnHelper = createColumnHelper<HistoryItem>();

const actionTranslate: Record<string, string> = {
    CreateTag: 'Create Tag',
    UpdateTag: 'Update Tag',
    DeleteTag: 'Delete Tag',
    AnalysisEnabledTag: 'Analysis Enabled Tag',
    AnalysisDisabledTag: 'Analysis Disabled Tag',
    CreateSelector: 'Create Selector',
    UpdateSelector: 'Update Selector',
    DeleteSelector: 'Delete Selector',
    CertifyNodeAuto: 'Automatic Certification',
    CertifyNodeManual: 'User Certification',
    CertifyNodeRevoked: 'Certify Revoked',
};

export const columns = [
    columnHelper.accessor('target', {
        id: 'target',
        header: () => {
            return <div className='pl-8 text-left'>Name</div>;
        },
        cell: ({ row }) => (
            <Tooltip tooltip={row.original.target} contentProps={{ align: 'start' }}>
                <div className='truncate'>{row.original.target}</div>
            </Tooltip>
        ),

        size: 150,
    }),
    columnHelper.accessor('action', {
        id: 'action',
        header: () => {
            return <div className='pl-8 text-left'>Action</div>;
        },
        cell: ({ row }) => (
            <Tooltip tooltip={actionTranslate[row.original.action]} contentProps={{ align: 'start' }}>
                <div className='truncate'>{actionTranslate[row.original.action]}</div>
            </Tooltip>
        ),
        size: 150,
    }),
    columnHelper.accessor('created_at', {
        id: 'created_at',
        header: () => {
            return <div className='pl-8 text-left'>Date</div>;
        },
        size: 96,
        cell: ({ row }) => <div>{DateTime.fromISO(row.original.created_at).toFormat(LuxonFormat.ISO_8601)}</div>,
    }),
    columnHelper.accessor('tagName', {
        id: 'tagName',
        header: () => {
            return <div className='pl-8 text-left'>Zone/Label</div>;
        },
        cell: ({ row }) => (
            <Tooltip tooltip={row.original.tagName || 'Deleted'} contentProps={{ align: 'start' }}>
                {row.original.tagName ? (
                    <div className='truncate'>{row.original.tagName}</div>
                ) : (
                    <div className='truncate italic'>Deleted</div>
                )}
            </Tooltip>
        ),
        size: 150,
    }),
    columnHelper.accessor('actor', {
        id: 'actor',
        header: () => {
            return <div className='pl-8 text-left'>Made By</div>;
        },
        cell: ({ row }: any) => (
            <Tooltip tooltip={row.original.email || row.original.actor} contentProps={{ align: 'start' }}>
                <div className='truncate'>{row.original.email || row.original.actor}</div>
            </Tooltip>
        ),
        size: 150,
    }),
    columnHelper.accessor('note', {
        id: 'note',
        header: () => {
            return <div className='pr-1'>Note</div>;
        },
        size: 96,
        cell: ({ row }) => <NoteCell row={row} />,
    }),
];
