// Copyright 2026 Specter Ops, Inc.
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

import { faInfoCircle } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { type ColumnDef } from '@tanstack/react-table';
import {
    Card,
    CardTitle,
    createColumnHelper,
    DataTable,
    IconButton,
    TableCell,
    TableRow,
    Tooltip,
    Typography,
} from 'doodle-ui';
import { type Extension } from 'js-client-library';
import { useState } from 'react';
import { SearchInput } from '../../components';
import { useDeleteExtension, useExtensionsQuery, usePermissions } from '../../hooks';
import { DEFAULT_NOTIFICATION, ERROR_NOTIFICATION, useNotifications } from '../../providers';
import { Permission } from '../../utils';
import { ConfirmDeleteExtensionDialog, DeleteExtensionButton } from './DeleteExtensionButton';

const columnHelper = createColumnHelper<Extension>();

export const ERROR_MESSAGE = 'There was an error fetching extensions';
export const LOADING_MESSAGE = 'Loading extensions...';
export const NO_DATA_MESSAGE = 'There are currently no active extensions';
export const NO_SEARCH_RESULTS_MESSAGE = 'No extensions match your search terms';

const TABLE_CELL_HEIGHT = 57;
const TABLE_HEADER_HEIGHT = 52;
const EMPTY_STATE_HEIGHT = TABLE_HEADER_HEIGHT + TABLE_CELL_HEIGHT * 2;

const TruncatedCell = ({ className = '', value }: { className?: string; value: string }) => {
    return (
        <Tooltip
            tooltip={value}
            contentProps={{
                align: 'start',
                'aria-hidden': true,
                className: 'min-[600px]:hidden dark:bg-neutral-5 dark:text-white border-neutral-300',
            }}>
            <div className={`${className} line-clamp-2 break-words leading-5`}>{value}</div>
        </Tooltip>
    );
};

export const ActiveExtensionsCard = () => {
    const [search, setSearch] = useState('');
    const [extensionToDelete, setExtensionToDelete] = useState<Extension | null>(null);
    const { data = [], isError, isLoading, isSuccess } = useExtensionsQuery();
    const deleteExtensionMutation = useDeleteExtension();
    const { addNotification } = useNotifications();
    const { checkPermission } = usePermissions();
    const hasDeletePermission = checkPermission(Permission.OPENGRAPH_WRITE);

    const handleDeleteClick = (extension: Extension) => {
        setExtensionToDelete(extension);
    };

    const handleDialogClose = () => {
        setExtensionToDelete(null);
    };

    const handleDelete = () => {
        if (!extensionToDelete) return;

        deleteExtensionMutation.mutate(extensionToDelete.id, {
            onSuccess: () => {
                addNotification(
                    `Extension "${extensionToDelete.name}" was deleted successfully!`,
                    'deleteExtensionSuccess',
                    DEFAULT_NOTIFICATION
                );
            },
            onError: () => {
                addNotification(
                    `Failed to delete extension "${extensionToDelete.name}". Please try again.`,
                    'deleteExtensionError',
                    ERROR_NOTIFICATION
                );
            },
            onSettled: handleDialogClose,
        });
    };

    const columns: ColumnDef<Extension, string>[] = [
        columnHelper.accessor('name', {
            id: 'name',
            size: 480,
            header: () => <span className='pl-6'>Name</span>,
            cell: ({ row }) => <TruncatedCell className='pl-6' value={row.original.name} />,
        }),
        columnHelper.accessor('namespace', {
            id: 'namespace',
            size: 240,
            header: () => (
                <div className='flex items-center gap-0.25'>
                    <span>Namespace</span>
                    <Tooltip
                        contentProps={{
                            className: 'max-w-96 dark:bg-neutral-5 dark:text-white border-neutral-300',
                        }}
                        tooltip={
                            <>
                                <Typography variant='caption' component='p'>
                                    Namespace Key is a set prefix for all node and edge kinds defined by an OpenGraph
                                    extension (e.g. GH_User, AWS_User).
                                </Typography>
                                <Typography variant='caption' component='p' className='mt-2'>
                                    This helps quickly inform which extension defines a node or edge kind and
                                    differentiate common types across platforms.
                                </Typography>
                            </>
                        }>
                        <IconButton
                            className='bg-transparent border-none p-0 cursor-default has-[svg]:p-0.5 hover:text-main dark:hover:text-main'
                            aria-label='Namespace information'
                            size={14}>
                            <FontAwesomeIcon icon={faInfoCircle} size='sm' />
                        </IconButton>
                    </Tooltip>
                </div>
            ),
            cell: ({ row }) => <TruncatedCell value={row.original.namespace} />,
        }),
        columnHelper.accessor('version', {
            id: 'version',
            size: 160,
            header: () => <span>Version</span>,
            cell: ({ row }) => <span>{row.original.version}</span>,
        }),
        columnHelper.display({
            id: 'delete-item',
            header: () => <span className='opacity-0'>Delete</span>,
            cell: ({ row }) => (
                <DeleteExtensionButton
                    extension={row.original}
                    onDeleteClick={handleDeleteClick}
                    hasDeletePermission={hasDeletePermission}
                />
            ),
            size: 56,
        }),
    ];

    const hasData = !isLoading && isSuccess && data.length > 0;
    const filteredData = data.filter((extension) => extension.name.toLowerCase().includes(search.toLowerCase()));
    const isEmptySearch = hasData && filteredData.length === 0;

    let fallbackMessage = LOADING_MESSAGE;

    if (isError) {
        fallbackMessage = ERROR_MESSAGE;
    } else if (isSuccess && !hasData) {
        fallbackMessage = NO_DATA_MESSAGE;
    } else if (isEmptySearch) {
        fallbackMessage = NO_SEARCH_RESULTS_MESSAGE;
    }

    return (
        <Card className='flex flex-col gap-4 overflow-hidden'>
            <header className='flex justify-between items-center pt-6 px-6 gap-3'>
                <CardTitle className='text-base'>Active Extensions</CardTitle>
                <SearchInput
                    className='self-start w-80'
                    id='search-active-extensions'
                    onInputChange={setSearch}
                    value={search}
                />
            </header>

            <div
                data-testid='active-extensions-table-container'
                style={{
                    height:
                        !hasData || isEmptySearch
                            ? `${EMPTY_STATE_HEIGHT}px`
                            : `${TABLE_HEADER_HEIGHT + TABLE_CELL_HEIGHT * filteredData.length}px`,
                }}>
                <DataTable
                    className='h-full'
                    data={filteredData}
                    noResultsFallback={
                        <TableRow>
                            <TableCell colSpan={4} className='h-28 text-center'>
                                {fallbackMessage}
                            </TableCell>
                        </TableRow>
                    }
                    columns={columns}
                    TableCellProps={{ className: 'h-[57px] py-2' }}
                    TableProps={{
                        className:
                            'table-fixed [&_tr>*:nth-child(1)]:!w-[42%] [&_tr>*:nth-child(2)]:!w-[28%] [&_tr>*:nth-child(3)]:!w-[16%] [&_tr>*:nth-child(4)]:!w-[14%]',
                        disableDefaultOverflowAuto: true,
                    }}
                />
            </div>

            <ConfirmDeleteExtensionDialog
                open={extensionToDelete !== null}
                extensionName={extensionToDelete?.name || ''}
                isDeleting={deleteExtensionMutation.isLoading}
                onAccept={handleDelete}
                onCancel={handleDialogClose}
            />
        </Card>
    );
};
