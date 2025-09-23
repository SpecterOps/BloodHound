import { Button } from '@bloodhoundenterprise/doodleui';
import {
    AssetGroupTagCertificationParams,
    AssetGroupTagCertificationRecord,
    CertificationManual,
    CertificationRevoked,
    CertificationType,
    UpdateCertificationRequest,
} from 'js-client-library';
import { FC, useCallback, useState } from 'react';
import { useInfiniteQuery, useMutation, useQuery } from 'react-query';
import { useParams } from 'react-router-dom';
import { DropdownOption, EntityInfoDataTable, EntityInfoPanel } from '../../../components';
import { useNotifications } from '../../../providers';
import { EntityKinds, apiClient } from '../../../utils';
import EntitySelectorsInformation from '../Details/EntitySelectorsInformation';
import CertificationTable from './CertificationTable';
import CertifyMembersConfirmDialog from './CertifyMembersConfirmDialog';

const Certification: FC = () => {
    const { tierId, labelId } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;
    const [search, setSearch] = useState('');
    const [filters, setFilters] = useState<AssetGroupTagCertificationParams>({});
    const [selectedRows, setSelectedRows] = useState<number[]>([]);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [certifyAction, setCertifyAction] = useState<typeof CertificationRevoked | typeof CertificationManual>(
        CertificationManual
    );

    const mockMemberId = 1;
    const PAGE_SIZE = 15;

    const useAssetGroupTagsCertificationsQuery = (filters?: AssetGroupTagCertificationParams, query?: string) => {
        const doSearch = query && query.length >= 3;
        const queryKey = doSearch ? query : 'static';
        return useInfiniteQuery<{
            count: number;
            data: { records: AssetGroupTagCertificationRecord[] };
            limit: number;
            skip: number;
        }>({
            queryKey: ['certifications', queryKey],
            queryFn: async ({ pageParam = 1 }) => {
                const skip = (pageParam - 1) * PAGE_SIZE;

                const result = doSearch
                    ? await apiClient.searchAssetGroupTagsCertifications(PAGE_SIZE, skip, query)
                    : await apiClient.getAssetGroupTagsCertifications(PAGE_SIZE, skip, filters);

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
        });
    };

    const memberQuery = useQuery({
        queryKey: ['zone-management', 'tags', tagId, 'member', mockMemberId],
        queryFn: async () => {
            if (!tagId || !mockMemberId) return undefined;
            return apiClient.getAssetGroupTagMemberInfo(tagId, mockMemberId).then((res) => {
                return res.data.data['member'];
            });
        },
        enabled: tagId !== undefined && mockMemberId !== undefined,
    });

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
            refetch();
        },
        onError: (error: any) => {
            console.log(error);
            addNotification('There was an error updating certification', `zone-management_update-certification_error`, {
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            });
        },
    });

    const { data, isLoading, isFetching, isSuccess, fetchNextPage, refetch } = useAssetGroupTagsCertificationsQuery(
        filters,
        search
    );

    const { addNotification } = useNotifications();

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

    const selectedNode = {
        id: memberQuery.data?.object_id,
        name: memberQuery.data?.name,
        type: memberQuery.data?.primary_kind as EntityKinds,
    };

    const showDialog = (action: typeof CertificationManual | typeof CertificationRevoked) => {
        setIsDialogOpen(true);
        setCertifyAction(action);
    };

    const filterByCertification = useCallback(
        (dropdownSelection: DropdownOption) => {
            const certificationStatus = dropdownSelection.key as CertificationType;
            filters.certificationStatus = certificationStatus;
            setFilters(filters);
            refetch();
        },
        [filters, refetch]
    );

    const handleConfirm = useCallback(
        (withNote: boolean, certifyNote?: string) => {
            setIsDialogOpen(false);
            const selectedMemberIds = selectedRows;
            if (selectedMemberIds.length === 0) {
                addNotification(
                    'Members must be selected for certification',
                    `zone-management_update-certification_no-members`,
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
                        <Button onClick={() => showDialog(CertificationManual)}>Certify</Button>
                        <Button variant='secondary' onClick={() => showDialog(CertificationRevoked)}>
                            Revoke
                        </Button>
                    </div>

                    <CertificationTable
                        data={data}
                        isLoading={isLoading}
                        isFetching={isFetching}
                        isSuccess={isSuccess}
                        fetchNextPage={fetchNextPage}
                        filterRows={filterByCertification}
                        selectedRows={selectedRows}
                        setSelectedRows={setSelectedRows}
                    />
                </div>
                <div className='basis-1/3'>
                    <div className='w-[400px] max-w-[400px]'>
                        <EntityInfoPanel
                            DataTable={EntityInfoDataTable}
                            selectedNode={selectedNode}
                            additionalTables={[
                                {
                                    sectionProps: { label: 'Selectors', id: memberQuery.data?.object_id },
                                    TableComponent: EntitySelectorsInformation,
                                },
                            ]}
                        />
                    </div>
                </div>
            </div>
            {isDialogOpen && <CertifyMembersConfirmDialog open={isDialogOpen} onConfirm={handleConfirm} />}
        </>
    );
};

export default Certification;
