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

import { Accordion, AccordionContent, AccordionItem, Button, Skeleton } from '@bloodhoundenterprise/doodleui';
import { faChevronUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { AssetGroupTagMemberListItem } from 'js-client-library';
import { useState } from 'react';
import { getListHeight } from '..';
import { SortableHeader } from '../../../components/ColumnHeaders';
import { InfiniteQueryFixedList, InfiniteQueryFixedListProps } from '../../../components/InfiniteQueryFixedList';
import NodeIcon from '../../../components/NodeIcon';
import { useRuleMembersInfiniteQuery, useTagMembersInfiniteQuery } from '../../../hooks/useAssetGroupTags';
import { useEnvironmentIdList } from '../../../hooks/useEnvironmentIdList';
import { usePZPathParams } from '../../../hooks/usePZParams/usePZPathParams';
import { privilegeZonesPath } from '../../../routes';
import { SortOrder } from '../../../types';
import { cn, useAppNavigate } from '../../../utils';

interface ObjectListsProps {
    kindCounts: Record<string, number>;
    totalCount: number;
}

export const ObjectsAccordion: React.FC<ObjectListsProps> = ({ kindCounts, totalCount }) => {
    const [openAccordion, setOpenAccordion] = useState('');

    return (
        <div>
            <div className='flex justify-between pl-4 pr-12 border-b border-neutral-3'>
                <span className='text-lg font-bold'>Objects</span>
                <span>
                    <span className='font-bold'>Total Objects:</span> {totalCount}
                </span>
            </div>
            <Accordion
                type='single'
                collapsible
                value={openAccordion}
                className='w-full min-w-0 rounded-none bg-neutral-2'
                data-testid='privilege-zones_details_objects-accordion'>
                {Object.entries(kindCounts)
                    .sort((a, b) => a[0].localeCompare(b[0]))
                    .map(([kind, count]) => (
                        <ObjectAccordionItem key={kind} kind={kind} count={count} onOpen={setOpenAccordion} />
                    ))}
            </Accordion>
        </div>
    );
};

interface ObjectAccordionItemProps {
    kind: string;
    count: number;
    onOpen: React.Dispatch<React.SetStateAction<string>>;
}

const LoadingRow = (_: number, style: React.CSSProperties) => (
    <div
        data-testid='privilege-zones_object-accordion_loading-skeleton'
        style={style}
        className='border-y border-neutral-3 relative w-full p-2'>
        <Skeleton className='h-full' />
    </div>
);

const ObjectAccordionItem: React.FC<ObjectAccordionItemProps> = ({ kind, count, onOpen }) => {
    const [sortOrder, setSortOrder] = useState<SortOrder>('asc');

    const navigate = useAppNavigate();

    const { ruleId, memberId, tagId, objectDetailsLink } = usePZPathParams();
    const environments = useEnvironmentIdList([{ path: `/${privilegeZonesPath}/*`, caseSensitive: false, end: false }]);

    const ruleMembersQuery = useRuleMembersInfiniteQuery(tagId, ruleId, sortOrder, environments, kind);
    const tagMembersQuery = useTagMembersInfiniteQuery(tagId, sortOrder, environments, kind);

    const handleClick = (id: number) => navigate(objectDetailsLink(tagId, id));

    const Row: InfiniteQueryFixedListProps<AssetGroupTagMemberListItem>['renderRow'] = (item, index, style) => {
        return (
            <div
                key={index}
                role='listitem'
                className={cn('pl-6 border-y border-neutral-3 relative', {
                    'bg-neutral-4': memberId === item.id.toString(),
                })}
                style={style}>
                <Button
                    variant='text'
                    title={`Type: ${item.primary_kind}; Name: ${item.name}`}
                    onClick={() => {
                        handleClick(item.id);
                    }}>
                    <span className='text-base text-contrast ml-2 truncate'>{item.name}</span>
                </Button>
            </div>
        );
    };

    return (
        <AccordionItem
            key={kind}
            value={kind}
            data-testid={`privilege-zones_details_${kind}-accordion-item`}
            className='[&[data-state=open]>div>div>button>svg]:rotate-180 sticky'>
            <div className='w-full flex items-center justify-between border-b border-neutral-3'>
                <div className='flex items-center gap-2'>
                    <Button
                        className='w-6'
                        variant='text'
                        onClick={() => {
                            onOpen((prev) => (prev === kind ? '' : kind));
                        }}>
                        <FontAwesomeIcon icon={faChevronUp} size='sm' className='font-bold' />
                    </Button>
                    <div className='flex items-center gap'>
                        <NodeIcon nodeType={kind} />
                        <SortableHeader
                            title={kind}
                            onSort={() => {
                                setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
                            }}
                            sortOrder={sortOrder}
                            classes={{
                                button: 'font-bold text-base',
                            }}
                        />
                    </div>
                </div>
                <span className='mr-12'>{count}</span>
            </div>
            <AccordionContent className='bg-neutral-2 p-0'>
                <div
                    className={cn(`border-neutral-5`, {
                        'h-[760px]': getListHeight(window.innerHeight) === 760,
                        'h-[640px]': getListHeight(window.innerHeight) === 640,
                        'h-[436px]': getListHeight(window.innerHeight) === 436,
                    })}>
                    <InfiniteQueryFixedList<AssetGroupTagMemberListItem>
                        itemSize={40}
                        queryResult={ruleId ? ruleMembersQuery : tagMembersQuery}
                        renderRow={Row}
                        renderLoadingRow={LoadingRow}
                    />
                </div>
            </AccordionContent>
        </AccordionItem>
    );
};
