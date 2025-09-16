import { AssetGroupTagCertificationRecord } from 'js-client-library';
import { FC, useState } from 'react';
import { useInfiniteQuery, useQuery } from 'react-query';
import { useParams } from 'react-router-dom';
import { EntityInfoDataTable, EntityInfoPanel } from '../../../components';
import { EntityKinds, apiClient } from '../../../utils';
import EntitySelectorsInformation from '../Details/EntitySelectorsInformation';
import CertificationTable from './CertificationTable';

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

    const selectedNode = {
        id: memberQuery.data?.object_id,
        name: memberQuery.data?.name,
        type: memberQuery.data?.primary_kind as EntityKinds,
    };

    return (
        <div className='flex gap-8 mt-4'>
            <div className='basis-2/3'>
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
    );
};

export default Certification;
