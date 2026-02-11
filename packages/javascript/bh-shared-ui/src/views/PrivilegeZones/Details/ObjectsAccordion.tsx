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

import { Accordion, AccordionContent, AccordionItem, Button, Skeleton, Tooltip } from '@bloodhoundenterprise/doodleui';
import { faChevronUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { AssetGroupTagMemberListItem } from 'js-client-library';
import { useState } from 'react';
import { SortableHeader } from '../../../components/ColumnHeaders';
import { InfiniteQueryFixedList, InfiniteQueryFixedListProps } from '../../../components/InfiniteQueryFixedList';
import NodeIcon from '../../../components/NodeIcon';
import { useRuleMembersInfiniteQuery, useTagMembersInfiniteQuery } from '../../../hooks/useAssetGroupTags';
import { useEnvironmentIdList } from '../../../hooks/useEnvironmentIdList';
import { ENVIRONMENT_AGGREGATION_SUPPORTED_ROUTES } from '../../../routes';
import { SortOrder } from '../../../types';
import { cn } from '../../../utils';
import { SelectedHighlight } from './SelectedHighlight';

export interface ObjectsAccordionProps {
    kindCounts: Record<string, number>;
    totalCount: number;
    tagId: string;
    ruleId?: string;
    objectId?: string;
    onObjectClick: (object: AssetGroupTagMemberListItem) => void;
}

export const ObjectsAccordion: React.FC<ObjectsAccordionProps> = ({
    ruleId,
    tagId,
    objectId,
    kindCounts,
    totalCount,
    onObjectClick,
}) => {
    const [openAccordion, setOpenAccordion] = useState('');

    return (
        <div>
            <div className='flex justify-between pl-4 pr-12 border-b border-neutral-3'>
                <span className='text-lg font-bold'>Objects</span>
                <span>
                    <span className='font-bold'>Total Objects:</span> {totalCount.toLocaleString()}
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
                        <ObjectAccordionItem
                            key={kind}
                            kind={kind}
                            count={count}
                            tagId={tagId}
                            ruleId={ruleId}
                            objectId={objectId}
                            isOpen={kind === openAccordion}
                            onOpen={setOpenAccordion}
                            onObjectClick={onObjectClick}
                        />
                    ))}
            </Accordion>
        </div>
    );
};

interface ObjectAccordionItemProps {
    kind: string;
    count: number;
    isOpen: boolean;
    tagId: string;
    ruleId?: string;
    objectId?: string;
    onOpen: React.Dispatch<React.SetStateAction<string>>;
    onObjectClick: (object: AssetGroupTagMemberListItem) => void;
}

const LoadingRow = (_: number, style: React.CSSProperties) => (
    <div
        data-testid='privilege-zones_object-accordion_loading-skeleton'
        style={style}
        className='border-y border-neutral-3 relative w-full p-2'>
        <Skeleton className='h-full' />
    </div>
);

const ObjectAccordionItem: React.FC<ObjectAccordionItemProps> = ({
    kind,
    count,
    tagId,
    ruleId,
    objectId,
    isOpen,
    onOpen,
    onObjectClick,
}) => {
    const [sortOrder, setSortOrder] = useState<SortOrder>('asc');

    const environments = useEnvironmentIdList(ENVIRONMENT_AGGREGATION_SUPPORTED_ROUTES, false);

    const ruleMembersQuery = useRuleMembersInfiniteQuery(tagId, ruleId, sortOrder, environments, kind, isOpen);
    const tagMembersQuery = useTagMembersInfiniteQuery(tagId, sortOrder, environments, kind, isOpen);

    const Row: InfiniteQueryFixedListProps<AssetGroupTagMemberListItem>['renderRow'] = (item, index, style) => {
        return (
            <Tooltip
                tooltip={<span className='text-contrast'>{item.name}</span>}
                contentProps={{ className: 'bg-neutral-3' }}>
                <div
                    key={index}
                    role='listitem'
                    className={cn('border-y border-neutral-3 relative', {
                        'bg-neutral-4': objectId === item.id.toString(),
                    })}
                    style={style}>
                    <SelectedHighlight itemId={item.id} type='member' />
                    <Button
                        variant='text'
                        className='w-full block text-left truncate'
                        onClick={() => {
                            onObjectClick(item);
                        }}>
                        <span className='pl-6 text-base text-contrast ml-2'>{item.name}</span>
                    </Button>
                </div>
            </Tooltip>
        );
    };

    return (
        <AccordionItem
            key={kind}
            value={kind}
            data-testid={`privilege-zones_details_${kind}-accordion-item`}
            className='[&[data-state=open]>div>div>button>svg]:rotate-180 sticky'>
            <div className='w-full flex items-center justify-between border-y border-neutral-3'>
                <div className='w-full flex items-center gap-2'>
                    <Button
                        className='w-6'
                        variant='text'
                        data-testid={`privilege-zones_details_${kind}-accordion_open-toggle-button`}
                        onClick={() => {
                            onOpen((prev) => (prev === kind ? '' : kind));
                        }}>
                        <FontAwesomeIcon icon={faChevronUp} size='sm' className='font-bold' />
                    </Button>
                    <div className='flex flex-1 items-center gap'>
                        <NodeIcon nodeType={kind} />
                        <SortableHeader
                            title={kind}
                            onSort={() => {
                                setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
                            }}
                            sortOrder={sortOrder}
                            classes={{
                                container: cn('flex-1', { 'pointer-events-none cursor-default': !isOpen }),
                                button: cn('font-bold text-base', {
                                    '[&>svg]:hidden': !isOpen,
                                }),
                            }}
                        />
                    </div>
                </div>
                <span className='mr-12'>{count.toLocaleString()}</span>
            </div>
            <AccordionContent className='bg-neutral-2 p-0'>
                <div className='border-neutral-5'>
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
