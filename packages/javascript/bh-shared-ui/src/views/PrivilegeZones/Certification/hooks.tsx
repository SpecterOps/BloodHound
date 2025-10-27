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

import { Checkbox, Tooltip, createColumnHelper } from '@bloodhoundenterprise/doodleui';
import { AssetGroupTagCertificationRecord } from 'js-client-library';
import { DateTime } from 'luxon';
import { OptionsObject } from 'notistack';
import { useMemo } from 'react';
import { useInfiniteQuery } from 'react-query';
import { useNotifications } from '../../..';
import { NodeIcon } from '../../../components';
import { privilegeZonesKeys, useAssetGroupTags, useAvailableEnvironments } from '../../../hooks';
import { LuxonFormat, apiClient } from '../../../utils';
import { ExtendedCertificationFilters } from './types';

const PAGE_SIZE = 50;

const createCertificationReqParams = (
    skip: number,
    limit: number,
    filters?: ExtendedCertificationFilters,
    search?: string
) => {
    const params = new URLSearchParams();
    params.append('skip', skip.toString());
    params.append('limit', PAGE_SIZE.toString());

    if (search) {
        // todo once API supports more fuzzy searching, name must be uppercase to work
        params.append('name', '~eq:' + search.toUpperCase());
    }

    if (!filters) return params;

    if (filters.certificationStatus !== undefined) params.append('certified', `eq:${filters.certificationStatus}`);

    if (filters.objectType) params.append('primary_kind', `eq:${filters.objectType}`);

    if (filters.approvedBy) params.append('certified_by', `eq:${filters.approvedBy}`);

    if (filters.tagId) params.append('asset_group_tag_id', `eq:${filters.tagId}`);

    if (filters['start-date'] && filters['start-date'] !== '') {
        params.append(
            'created_at',
            'gte:' + DateTime.fromFormat(filters['start-date'], LuxonFormat.ISO_8601).startOf('day').toISO()
        );
    }

    if (filters['end-date'] && filters['end-date'] !== '') {
        params.append(
            'created_at',
            'lte:' + DateTime.fromFormat(filters['end-date'], LuxonFormat.ISO_8601).endOf('day').toISO()
        );
    }

    return params;
};

export const useAssetGroupTagsCertificationsQuery = (filters?: ExtendedCertificationFilters, search?: string) => {
    return useInfiniteQuery<{
        count: number;
        data: { members: AssetGroupTagCertificationRecord[] };
        limit: number;
        skip: number;
    }>({
        queryKey: privilegeZonesKeys.certifications(filters, search),
        queryFn: async ({ pageParam = 1 }) => {
            const skip = (pageParam - 1) * PAGE_SIZE;

            const params = createCertificationReqParams(skip, PAGE_SIZE, filters, search);

            const result = await apiClient.getAssetGroupTagsCertifications({ params });

            return result.data;
        },
        getNextPageParam: (lastPage) => {
            const nextPage = lastPage.skip / PAGE_SIZE + 2;

            if ((nextPage - 1) * PAGE_SIZE >= lastPage.count) {
                return undefined;
            }

            return nextPage;
        },
        getPreviousPageParam: (firstPage) => {
            if (firstPage.skip === 0) {
                return undefined;
            }

            return firstPage.skip / PAGE_SIZE - 1;
        },
        keepPreviousData: false,
    });
};

const columnHelper = createColumnHelper<AssetGroupTagCertificationRecord>();

export const useCertificationColumns = ({
    onRowSelect,
    toggleAllRowsSelected,
    selectedRows,
    allRowsAreSelected,
}: {
    onRowSelect: (row: AssetGroupTagCertificationRecord) => void;
    toggleAllRowsSelected: () => void;
    selectedRows: Record<string, boolean>;
    allRowsAreSelected: boolean;
}) => {
    const { data: availableEnvironments = [] } = useAvailableEnvironments();
    const { data: assetGroupTags = [] } = useAssetGroupTags();

    const ids = Object.keys(selectedRows);

    const environmentMap = useMemo(() => {
        const map = new Map<string, string>();
        for (const environment of availableEnvironments) {
            map.set(environment.id, environment.name);
        }
        return map;
    }, [availableEnvironments]);

    const tagMap = useMemo(() => {
        const map = new Map<number, string>();
        for (const tag of assetGroupTags) {
            map.set(tag.id, tag.name);
        }
        return map;
    }, [assetGroupTags]);

    const columns = useMemo(() => {
        return [
            columnHelper.display({
                id: 'bulk_certify',
                header: () => {
                    const checked = allRowsAreSelected ? true : ids.length > 0 ? 'indeterminate' : false;

                    return (
                        <div className='flex justify-center'>
                            <Checkbox
                                data-testid='certification-table-select-all'
                                checked={checked}
                                onCheckedChange={() => {
                                    toggleAllRowsSelected();
                                }}
                            />
                        </div>
                    );
                },
                cell: (info) => (
                    <div className='flex justify-center'>
                        <Checkbox
                            onClick={(e) => {
                                e.stopPropagation();
                            }}
                            data-testid={`certification-table-row-${info.row.original.id}`}
                            checked={ids.includes(info.row.original.id.toString())}
                            onCheckedChange={() => {
                                onRowSelect(info.row.original);
                            }}
                        />
                    </div>
                ),
                size: 48,
            }),
            columnHelper.accessor('primary_kind', {
                header: 'Type',
                cell: (info) => <div className='ml-2'>{<NodeIcon nodeType={info.getValue()} />}</div>,
                size: 36,
            }),
            columnHelper.accessor('name', {
                header: 'Member Name',
                cell: (info) => {
                    return (
                        <Tooltip tooltip={info.getValue()}>
                            <div className='min-w-0 truncate'>{info.getValue()}</div>
                        </Tooltip>
                    );
                },
                size: 128,
            }),
            columnHelper.accessor('environment_id', {
                header: 'Environment',
                cell: (info) => (
                    <div className='min-w-0 truncate'>{environmentMap.get(info.getValue()) ?? 'Unknown'}</div>
                ),
                size: 72,
            }),
            columnHelper.accessor('asset_group_tag_id', {
                header: 'Zone',
                cell: (info) => <div className='min-w-0 truncate'>{tagMap.get(info.getValue()) ?? 'Unknown'}</div>,
                size: 72,
            }),
            columnHelper.accessor('created_at', {
                header: 'First Seen',
                cell: (info) => (
                    <div className='text-left'>{DateTime.fromISO(info.getValue()).toFormat(LuxonFormat.ISO_8601)}</div>
                ),
                size: 48,
            }),
        ];
    }, [environmentMap, tagMap, onRowSelect, toggleAllRowsSelected]);

    return columns;
};

const notificationOptions: OptionsObject = {
    anchorOrigin: { vertical: 'top', horizontal: 'right' },
};

export const useCertificationNotifications = () => {
    const { addNotification } = useNotifications();

    const certificationSuccess = () =>
        addNotification(
            'Selected Certification Successful',
            'privilege-zones_update-certification_success',
            notificationOptions
        );

    const revocationSuccess = () =>
        addNotification(
            'Selected Revocation Successful',
            'privilege-zones_update-revocation_success',
            notificationOptions
        );

    const updateError = () =>
        addNotification(
            'There was an error updating certification',
            'privilege-zones_update-certification_error',
            notificationOptions
        );

    const noRowsSelectedError = () =>
        addNotification(
            'Members must be selected for certification',
            'privilege_zones_update-certification_no-members',
            notificationOptions
        );

    return { certificationSuccess, revocationSuccess, updateError, noRowsSelectedError };
};
