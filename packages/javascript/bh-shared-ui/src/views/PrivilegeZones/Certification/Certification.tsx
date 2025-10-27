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

import { Button, Card, CardHeader, CardTitle } from '@bloodhoundenterprise/doodleui';
import {
    AssetGroupTagCertificationRecord,
    CertificationManual,
    CertificationRevoked,
    CertificationType,
    CertificationTypeMap,
    UpdateCertificationRequest,
} from 'js-client-library';
import { useCallback, useEffect, useState } from 'react';
import { useMutation, useQueryClient } from 'react-query';
import { useSearchParams } from 'react-router-dom';
import { SelectedNode, privilegeZonesKeys, useExploreParams, useMemberInfo, usePZQueryParams } from '../../..';
import { AppIcon, DropdownOption, DropdownSelector, EntityInfoDataTable, EntityInfoPanel } from '../../../components';
import { SearchInput } from '../../../components/SearchInput';
import { zonesPath } from '../../../routes';
import { EntityKinds, apiClient } from '../../../utils';
import EntitySelectorsInformation from '../Details/EntitySelectorsInformation';
import CertificationTable from './CertificationTable';
import CertifyMembersConfirmDialog from './CertifyMembersConfirmDialog';
import FilterDialog from './FilterDialog';
import { certOptions, certificationCountTextMap, defaultFilterValues, emptyPaginatedData } from './constants';
import { useAssetGroupTagsCertificationsQuery, useCertificationNotifications } from './hooks';
import { ExtendedCertificationFilters } from './types';

const allSelectedRowsFromItems = (items: AssetGroupTagCertificationRecord[]) => {
    return items.reduce((accumulator: Record<string, boolean>, currentValue) => {
        accumulator[currentValue.id.toString()] = true;
        return accumulator;
    }, {});
};

const emptyParams = new URLSearchParams();

type CertifyAction = typeof CertificationManual | typeof CertificationRevoked;

const Certification = () => {
    const [search, setSearch] = useState('');
    const [filters, setFilters] = useState<ExtendedCertificationFilters>(defaultFilterValues);
    const [selectedRows, setSelectedRows] = useState<Record<string, boolean>>({});
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [dropdownSelection, setDropdownSelection] =
        useState<(typeof CertificationTypeMap)[CertificationType]>('Pending');
    const [action, setAction] = useState<CertifyAction>(CertificationManual);

    const [searchParams, setSearchParams] = useSearchParams();

    const { selectedItem } = useExploreParams();
    const { assetGroupTagId } = usePZQueryParams();
    const memberQuery = useMemberInfo(assetGroupTagId?.toString(), selectedItem ?? '');
    const selectedNode: SelectedNode | null = memberQuery.data
        ? {
              id: memberQuery.data.object_id,
              name: memberQuery.data.name,
              type: memberQuery.data.primary_kind as EntityKinds,
          }
        : null;

    const certificationsQuery = useAssetGroupTagsCertificationsQuery(filters, search);
    const certificationsData = certificationsQuery.data ?? emptyPaginatedData;
    const certificationItems = certificationsData.pages.flatMap((page) => page.data?.members ?? []);
    const count = certificationsData.pages[0].count;

    const setURLParamsForEntityInfo = useCallback(
        (id: string, assetGroupTagId?: string) => {
            const params = new URLSearchParams(searchParams);
            params.set('selectedItem', id);
            assetGroupTagId && params.set('assetGroupTagId', assetGroupTagId);
            params.set('expandedPanelSections', 'Selectors');
            setSearchParams(params);
        },
        [setSearchParams, searchParams]
    );

    const showDialog = (action: CertifyAction) => {
        setIsDialogOpen(true);
        setAction(action);
    };

    const clearSelectedRows = () => setSelectedRows({});

    const clearSearchParams = useCallback(() => setSearchParams(emptyParams), [setSearchParams]);

    const addRowToSelectedRows = useCallback((id: string) => {
        setSelectedRows((prev) => {
            return { ...prev, [id]: true };
        });
    }, []);

    const removeRowFromSelectedRows = useCallback(
        (rowData: AssetGroupTagCertificationRecord) => {
            setSelectedRows((prev) => {
                const newRows = { ...prev };
                delete newRows[rowData.id.toString()];

                const ids = Object.keys(newRows);

                const oneRowLeftSelected = ids.length === 1;
                if (oneRowLeftSelected) setURLParamsForEntityInfo(ids[0], rowData.asset_group_tag_id.toString());

                return newRows;
            });
        },
        [setURLParamsForEntityInfo]
    );

    const { certificationSuccess, revocationSuccess, updateError, noRowsSelectedError } =
        useCertificationNotifications();
    const queryClient = useQueryClient();

    const certifyMutation = useMutation({
        mutationFn: async (requestBody: UpdateCertificationRequest) => {
            return apiClient.updateAssetGroupTagCertification(requestBody);
        },
        onSuccess: () => {
            action === CertificationManual ? certificationSuccess() : revocationSuccess();
            clearSelectedRows();
            queryClient.invalidateQueries({ queryKey: privilegeZonesKeys.certifications(filters, search) });
        },
        onError: (error: any) => {
            console.error(error);
            updateError();
        },
    });

    const filterByCertification = useCallback((dropdownSelection: DropdownOption) => {
        const certificationStatus = dropdownSelection.key as CertificationType;
        setFilters((prev) => ({ ...prev, certificationStatus }));
    }, []);

    const handleConfirm = useCallback(
        (note?: string) => {
            setIsDialogOpen(false);
            const ids = Object.keys(selectedRows).map((id) => parseInt(id));
            const noRowsAreSelected = ids.length === 0;

            if (noRowsAreSelected) {
                noRowsSelectedError();
                return;
            }

            certifyMutation.mutate({
                member_ids: ids,
                action,
                note,
            });
        },
        [noRowsSelectedError, action, certifyMutation, selectedRows]
    );

    const handleRowCheck = useCallback(
        (row: AssetGroupTagCertificationRecord) => {
            const ids = Object.keys(selectedRows);
            const someRowsAreSelected = ids.length > 0;

            if (someRowsAreSelected) {
                const clickedRowIsAlreadySelected = ids.includes(row.id.toString());

                clickedRowIsAlreadySelected ? removeRowFromSelectedRows(row) : addRowToSelectedRows(row.id.toString());
                return;
            }

            addRowToSelectedRows(row.id.toString());
            setURLParamsForEntityInfo(row.id.toString(), row.asset_group_tag_id.toString());
        },
        [selectedRows, addRowToSelectedRows, setURLParamsForEntityInfo, removeRowFromSelectedRows]
    );

    const handleRowSelect = useCallback(
        (row: AssetGroupTagCertificationRecord) => {
            setURLParamsForEntityInfo(row.id.toString(), row.asset_group_tag_id.toString());
        },
        [setURLParamsForEntityInfo]
    );

    const toggleAllRowsSelected = useCallback(() => {
        const ids = Object.keys(selectedRows);
        const allRowsAreSelected = ids.length > 0 && ids.length === certificationItems.length;

        allRowsAreSelected ? clearSelectedRows() : setSelectedRows(allSelectedRowsFromItems(certificationItems));
    }, [selectedRows, setSelectedRows, certificationItems]);

    useEffect(() => {
        // clear selection whenever the dropdown filter changes
        clearSelectedRows();
    }, [dropdownSelection, setSelectedRows]);

    return (
        <div className='grow'>
            <div className='flex gap-4 my-4'>
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

            <div className='flex gap-4 h-[75dvh] grow'>
                <Card className='grow'>
                    <CardHeader className='pl-8'>
                        <CardTitle>
                            <div className='flex justify-between'>
                                <span>
                                    Certifications
                                    <span className='text-sm font-normal ml-2'>{`${count} ${certificationCountTextMap[dropdownSelection] ?? 'pending'}`}</span>
                                </span>
                                <div className='flex font-normal'>
                                    <SearchInput value={search} onInputChange={setSearch} />
                                    <FilterDialog filters={filters} setFilters={setFilters} />
                                </div>
                            </div>
                        </CardTitle>
                        <div className='pb-2'>
                            <DropdownSelector
                                variant='transparent'
                                options={certOptions}
                                selectedText={
                                    <div className='flex items-center gap-3'>
                                        <AppIcon.CertStatus size={24} /> <p>{`${dropdownSelection}`}</p>
                                    </div>
                                }
                                onChange={(selectedCertificationType: DropdownOption) => {
                                    setDropdownSelection(selectedCertificationType.value);
                                    filterByCertification(selectedCertificationType);
                                }}
                            />
                        </div>
                    </CardHeader>
                    <CertificationTable
                        query={certificationsQuery}
                        onRowSelect={handleRowSelect}
                        onRowCheck={handleRowCheck}
                        toggleAllRowsSelected={toggleAllRowsSelected}
                        selectedRows={selectedRows}
                        items={certificationItems}
                        count={count}
                    />
                </Card>

                <div className='w-[400px] min-w-[400px] overflow-y-auto'>
                    <EntityInfoPanel
                        DataTable={EntityInfoDataTable}
                        selectedNode={selectedNode}
                        priorityTables={[
                            {
                                sectionProps: {
                                    tagType: zonesPath,
                                    tagId: memberQuery.data?.asset_group_tag_id,
                                },
                                TableComponent: EntitySelectorsInformation,
                            },
                        ]}
                    />
                </div>
            </div>
            {isDialogOpen && (
                <CertifyMembersConfirmDialog
                    open={isDialogOpen}
                    onClose={() => setIsDialogOpen(false)}
                    onConfirm={handleConfirm}
                />
            )}
        </div>
    );
};

export default Certification;
