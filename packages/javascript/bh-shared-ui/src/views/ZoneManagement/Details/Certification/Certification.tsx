import { FC } from 'react';
import { useQuery } from 'react-query';
import { useParams } from 'react-router-dom';
import { EntityInfoDataTable, EntityInfoPanel } from '../../../../components';
import { EntityKinds, apiClient } from '../../../../utils';
import EntitySelectorsInformation from '../EntitySelectorsInformation';
import CertificationTable from './CertificationTable';

const Certification: FC = () => {
    const { tierId, labelId } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;

    const mockMemberId = 1;

    const mockObjectData = [
        {
            type: 1,
            name: 'object 1',
            environment: 'env',
            zone: '1',
            first_seen: '',
            certified: 'AUTO',
            certified_by: 'test@specterops.io',
            object_id: '1',
            primary_kind: 'User',
        },
        {
            type: 1,
            name: 'object 2',
            environment: 'env',
            zone: '1',
            first_seen: '',
            certified: 'AUTO',
            certified_by: 'test@specterops.io',
            object_id: '2',
            primary_kind: 'User',
        },
        {
            type: 1,
            name: 'object 3',
            environment: 'env',
            zone: '1',
            first_seen: '',
            certified: 'AUTO',
            certified_by: 'test@specterops.io',
            object_id: '3',
            primary_kind: 'User',
        },
    ];

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

    const selectedNode = {
        id: memberQuery.data?.object_id,
        name: memberQuery.data?.name,
        type: memberQuery.data?.primary_kind as EntityKinds,
    };

    return (
        <div className='flex gap-8 mt-4'>
            <div className='basis-2/3'>
                <CertificationTable />
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
