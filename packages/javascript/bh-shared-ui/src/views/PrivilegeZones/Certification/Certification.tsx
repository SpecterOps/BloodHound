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
import { Button } from '@bloodhoundenterprise/doodleui';
import {
    AssetGroupTagCertificationRecord,
    CertificationManual,
    CertificationRevoked,
    CertificationType,
    CertificationTypeMap,
    UpdateCertificationRequest,
} from 'js-client-library';
import { DateTime } from 'luxon';
import { useCallback, useState } from 'react';
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from 'react-query';
import { DropdownOption, EntityInfoDataTable, EntityInfoPanel } from '../../../components';
import { useNotifications } from '../../../providers';
import { zonesPath } from '../../../routes';
import { SelectedNode } from '../../../types';
import { EntityKinds, LuxonFormat, apiClient } from '../../../utils';
import EntitySelectorsInformation from '../Details/EntitySelectorsInformation';
import CertificationTable from './CertificationTable';
import CertifyMembersConfirmDialog from './CertifyMembersConfirmDialog';
import { defaultFilterValues } from './constants';
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

    if (!filters) return params;

    if (filters.certificationStatus !== undefined) params.append('certified', `eq:${filters.certificationStatus}`);

    if (filters.objectType) params.append('primary_kind', `eq:${filters.objectType}`);

    if (filters.approvedBy) params.append('certified_by', `eq:${filters.approvedBy}`);

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

    if (search) {
        // todo once API supports more fuzzy searching, name must be uppercase to work
        params.append('name', '~eq:' + search.toUpperCase());
    }

    return params;
};

const useAssetGroupTagsCertificationsQuery = (filters?: ExtendedCertificationFilters, search?: string) => {
    return useInfiniteQuery<{
        count: number;
        data: { members: AssetGroupTagCertificationRecord[] };
        limit: number;
        skip: number;
    }>({
        queryKey: ['certifications', filters, search],
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

const Certification = () => {
    const [selectedTagId, setSelectedTagId] = useState<number | undefined>(undefined);
    const [selectedMemberId, setSelectedMemberId] = useState<number | undefined>(undefined);
    const [search, setSearch] = useState('');
    const [filters, setFilters] = useState<ExtendedCertificationFilters>(defaultFilterValues);
    const [selectedRows, setSelectedRows] = useState<number[]>([]);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [dropdownSelection, setDropdownSelection] =
        useState<(typeof CertificationTypeMap)[CertificationType]>('Pending');
    const [certifyAction, setCertifyAction] = useState<typeof CertificationRevoked | typeof CertificationManual>(
        CertificationManual
    );

    const queryClient = useQueryClient();

    const { addNotification } = useNotifications();

    const certifyMutation = useMutation({
        mutationFn: async (requestBody: UpdateCertificationRequest) => {
            return apiClient.updateAssetGroupTagCertification(requestBody);
        },
        onSuccess: () => {
            const certificationType = certifyAction === CertificationManual ? 'Certification' : 'Revocation';
            addNotification(
                `Selected ${certificationType} Successful`,
                `zone-management_update-certification_success`,
                {
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                }
            );
            setSelectedRows([]);
            queryClient.invalidateQueries({ queryKey: ['certifications', filters] });
        },
        onError: (error: any) => {
            console.error(error);
            addNotification('There was an error updating certification', `zone-management_update-certification_error`, {
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            });
        },
    });

    const { data, isLoading, isFetching, isSuccess, fetchNextPage } = useAssetGroupTagsCertificationsQuery(
        filters,
        search
    );

    const memberQuery = useQuery({
        queryKey: ['privilege-zones', 'tags', selectedTagId, 'member', selectedMemberId],
        queryFn: async () => {
            if (!selectedTagId || !selectedMemberId) return undefined;
            return apiClient.getAssetGroupTagMemberInfo(selectedTagId, selectedMemberId).then((res) => {
                return res.data.data['member'];
            });
        },
        enabled: !!selectedTagId && !!selectedMemberId,
    });

    const createCertificationRequestBody = (
        action: typeof CertificationManual | typeof CertificationRevoked,
        objectIds: number[],
        withNote: boolean,
        note?: string
    ): UpdateCertificationRequest => {
        return {
            member_ids: objectIds,
            action: action,
            note: withNote ? note : undefined,
        };
    };

    const selectedNode: SelectedNode | null = memberQuery.data
        ? {
              id: memberQuery.data.object_id?.toString() ?? '',
              name: memberQuery.data.name ?? 'Unknown',
              type: memberQuery.data.primary_kind as EntityKinds,
          }
        : null;

    const showDialog = (action: typeof CertificationManual | typeof CertificationRevoked) => {
        setIsDialogOpen(true);
        setCertifyAction(action);
    };

    const filterByCertification = useCallback((dropdownSelection: DropdownOption) => {
        const certificationStatus = dropdownSelection.key as CertificationType;
        setFilters((prev) => ({ ...prev, certificationStatus }));
    }, []);

    const applyAdvancedFilters = useCallback((advancedFilters: Partial<ExtendedCertificationFilters>) => {
        setFilters((prev) => ({ ...prev, ...advancedFilters }));
    }, []);

    const handleConfirm = useCallback(
        (withNote: boolean, certifyNote?: string) => {
            setIsDialogOpen(false);
            const selectedMemberIds = selectedRows;
            if (selectedMemberIds.length === 0) {
                addNotification(
                    'Members must be selected for certification',
                    `privilege_zones_update-certification_no-members`,
                    {
                        anchorOrigin: { vertical: 'top', horizontal: 'right' },
                    }
                );
                return;
            }
            const requestBody = createCertificationRequestBody(certifyAction, selectedMemberIds, withNote, certifyNote);
            certifyMutation.mutate(requestBody);
        },
        [addNotification, certifyAction, certifyMutation, selectedRows]
    );

    return (
        <>
            <div className='flex gap-8 mt-4'>
                <div className='basis-2/3'>
                    <div className='flex gap-4 mb-4'>
                        <Button
                            onClick={() => showDialog(CertificationManual)}
                            disabled={dropdownSelection === 'Automatic Certification'}>
                            Certify
                        </Button>
                        <Button
                            variant='secondary'
                            onClick={() => showDialog(CertificationRevoked)}
                            disabled={dropdownSelection === 'Automatic Certification'}>
                            Revoke
                        </Button>
                    </div>

                    <CertificationTable
                        data={data}
                        filters={filters}
                        setFilters={setFilters}
                        search={search}
                        setSearch={setSearch}
                        onRowSelect={(row) => {
                            setSelectedMemberId(row?.id);
                            setSelectedTagId(row?.asset_group_tag_id);
                        }}
                        isLoading={isLoading}
                        isFetching={isFetching}
                        isSuccess={isSuccess}
                        fetchNextPage={fetchNextPage}
                        filterRows={filterByCertification}
                        applyAdvancedFilters={applyAdvancedFilters}
                        selectedRows={selectedRows}
                        setSelectedRows={setSelectedRows}
                        dropdownSelection={dropdownSelection}
                        setDropdownSelection={setDropdownSelection}
                    />
                </div>
                <div className='basis-1/3'>
                    <div className='w-[400px] max-w-[400px]'>
                        {memberQuery.data && selectedNode ? (
                            <EntityInfoPanel
                                DataTable={EntityInfoDataTable}
                                selectedNode={selectedNode}
                                additionalTables={[
                                    {
                                        sectionProps: {
                                            memberId: memberQuery.data.id,
                                            tagType: zonesPath,
                                            tagId: memberQuery.data.asset_group_tag_id,
                                        },
                                        TableComponent: EntitySelectorsInformation,
                                    },
                                ]}
                            />
                        ) : (
                            <EntityInfoPanel DataTable={EntityInfoDataTable} selectedNode={null} />
                        )}
                    </div>
                </div>
            </div>
            {isDialogOpen && (
                <CertifyMembersConfirmDialog
                    open={isDialogOpen}
                    onClose={() => setIsDialogOpen(false)}
                    onConfirm={handleConfirm}
                />
            )}
        </>
    );
};

export default Certification;
