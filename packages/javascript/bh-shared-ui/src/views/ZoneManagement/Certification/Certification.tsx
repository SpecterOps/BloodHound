import {
    AssetGroupTagCertificationRecord,
    CertificationManual,
    CertificationRevoked,
    UpdateCertificationRequest,
} from 'js-client-library';
import { FC, useState } from 'react';
import { useInfiniteQuery, useMutation, useQuery } from 'react-query';
import { useParams } from 'react-router-dom';
import { EntityInfoDataTable, EntityInfoPanel } from '../../../components';
import { useNotifications } from '../../../providers';
import { EntityKinds, apiClient } from '../../../utils';
import EntitySelectorsInformation from '../Details/EntitySelectorsInformation';
import CertificationTable from './CertificationTable';
import CertifyMembersConfirmDialog from './CertifyMembersConfirmDialog';

const Certification: FC = () => {
    const { tierId, labelId } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;
    const [search, setSearch] = useState('');
    const [filters, setFilters] = useState();

    const mockMemberId = 1;
    const PAGE_SIZE = 15;

    const useAssetGroupTagsCertificationsQuery = (filters, query?: string) => {
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
                    : await apiClient.getAssetGroupTagsCertifications(PAGE_SIZE, skip);

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

    const { data, isLoading, isFetching, isSuccess, fetchNextPage } = useAssetGroupTagsCertificationsQuery(
        filters,
        search
    );

    const { addNotification } = useNotifications();

    const certifyMutation = useMutation({
        mutationFn: async (requestBody: UpdateCertificationRequest) => {
            return apiClient.updateAssetGroupTagCertification(requestBody);
        },
        onSuccess: () => {
            const certificationType = certifyOrRevoke === CertificationManual ? 'Certification' : 'Revocation';
            addNotification(`Selected ${certificationType} Successful`);
            // TODO soft-refresh the page, keeping the selected filters active
        },
        onError: (error: any) => {
            addNotification('There was an error updating certification.');
        },
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

    const getSelectedMemberIds = () => {
        //TODO unmock
        return [1, 2, 3];
    };

    const selectedNode = {
        id: memberQuery.data?.object_id,
        name: memberQuery.data?.name,
        type: memberQuery.data?.primary_kind as EntityKinds,
    };

    const showDialog = (choice: typeof CertificationManual | typeof CertificationRevoked) => {
        setIsDialogOpen(true);
        setCertifyOrRevoke(choice);
    };

    const [isDialogOpen, setIsDialogOpen] = useState(false);
    // TODO -- is there a way to do this without setting a default value??
    const [certifyOrRevoke, setCertifyOrRevoke] = useState<typeof CertificationRevoked | typeof CertificationManual>(
        CertificationManual
    );

    const handleConfirm = (withNote: boolean, certifyNote?: string) => {
        console.log('Confirmed! Note: ', certifyNote);
        setIsDialogOpen(false);

        const selectedMemberIds = getSelectedMemberIds();

        if (selectedMemberIds.length === 0) {
            addNotification('No objects selected');
            return;
        }
        // TODO -- input sanitization for the note
        const requestBody = createCertificationRequestBody(certifyOrRevoke, selectedMemberIds, withNote, certifyNote);
        // TODO -- make API call
        certifyMutation.mutate(requestBody);
    };

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
