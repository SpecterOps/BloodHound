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
import { AssetGroupTagSelector, CustomRulesKey, DefaultRulesKey, DisabledRulesKey, RulesKey } from 'js-client-library';
import { useState } from 'react';
import { SortableHeader } from '../../../components/ColumnHeaders';
import { InfiniteQueryFixedList, InfiniteQueryFixedListProps } from '../../../components/InfiniteQueryFixedList';
import { useRulesInfiniteQuery } from '../../../hooks/useAssetGroupTags';
import { useEnvironmentIdList } from '../../../hooks/useEnvironmentIdList';
import { usePZPathParams } from '../../../hooks/usePZParams/usePZPathParams';
import { useSelectedTagPathParams } from '../../../hooks/useSelectedTag';
import { privilegeZonesPath } from '../../../routes';
import { SortOrder, SortOrderAscending, SortOrderDescending } from '../../../types';
import { cn, useAppNavigate } from '../../../utils';

type RuleSection = typeof CustomRulesKey | typeof DefaultRulesKey | typeof DisabledRulesKey;

const filters = {
    [CustomRulesKey]: { isDefault: false, disabled: false },
    [DefaultRulesKey]: { isDefault: true, disabled: false },
    [DisabledRulesKey]: { disabled: true },
};

const filterLabels = {
    [CustomRulesKey]: 'Custom Rules',
    [DefaultRulesKey]: 'Default Rules',
    [DisabledRulesKey]: 'Disabled Rules',
} as const;

export const RulesAccordion: React.FC = () => {
    const [openAccordion, setOpenAccordion] = useState<RuleSection | ''>(CustomRulesKey);
    const selectedTag = useSelectedTagPathParams();
    const { ruleId, tagDetailsLink, tagId } = usePZPathParams();
    const navigate = useAppNavigate();

    if (!selectedTag.counts) return null;

    return (
        <div>
            <div className='flex justify-between pl-4 pr-12 border-b border-neutral-3'>
                <span className='text-lg font-bold'>Rules</span>
                <span>
                    <span className='font-bold'>Total Rules:</span> {selectedTag.counts[RulesKey]}
                </span>
            </div>
            <div
                className={cn('pl-6 border-y border-neutral-3 relative', {
                    'bg-neutral-4': !ruleId,
                })}>
                <Button variant='text' onClick={() => navigate(tagDetailsLink(tagId))}>
                    <span className='text-base text-contrast ml-2 truncate'>All Rules</span>
                </Button>
            </div>
            <Accordion
                type='single'
                collapsible
                value={openAccordion}
                className='w-full min-w-0 rounded-none bg-neutral-2'
                data-testid='privilege-zones_details_rules-accordion'>
                <RuleAccordionItem
                    section={CustomRulesKey}
                    count={selectedTag.counts[CustomRulesKey]}
                    onOpen={setOpenAccordion}
                />
                <RuleAccordionItem
                    section={DefaultRulesKey}
                    count={selectedTag.counts[DefaultRulesKey]}
                    onOpen={setOpenAccordion}
                />
                <RuleAccordionItem
                    section={DisabledRulesKey}
                    count={selectedTag.counts[DisabledRulesKey]}
                    onOpen={setOpenAccordion}
                />
            </Accordion>
        </div>
    );
};

interface RuleAccordionItemProps {
    section: RuleSection;
    count: number;
    onOpen: React.Dispatch<React.SetStateAction<RuleSection | ''>>;
}

const LoadingRow = (_: number, style: React.CSSProperties) => (
    <div
        data-testid='privilege-zones_rule-accordion_loading-skeleton'
        style={style}
        className='border-y border-neutral-3 relative w-full p-2'>
        <Skeleton className='h-full' />
    </div>
);

const RuleAccordionItem: React.FC<RuleAccordionItemProps> = ({ section: filterKey, count, onOpen }) => {
    const [sortOrder, setSortOrder] = useState<SortOrder>(SortOrderAscending);

    const navigate = useAppNavigate();

    const { ruleId, tagId, ruleDetailsLink } = usePZPathParams();
    const environments = useEnvironmentIdList([{ path: `/${privilegeZonesPath}/*`, caseSensitive: false, end: false }]);

    const rulesQuery = useRulesInfiniteQuery(tagId, { sortOrder, environments, ...filters[filterKey] });

    const handleClick = (id: number) => navigate(ruleDetailsLink(tagId, id));

    const Row: InfiniteQueryFixedListProps<AssetGroupTagSelector>['renderRow'] = (item, index, style) => {
        return (
            <div
                key={index}
                role='listitem'
                className={cn('pl-6 border-y border-neutral-3 relative', {
                    'bg-neutral-4': ruleId === item.id.toString(),
                })}
                style={style}>
                <Button variant='text' title={`Name: ${item.name}`} onClick={() => handleClick(item.id)}>
                    <span className='text-base text-contrast ml-2 truncate'>{item.name}</span>
                </Button>
            </div>
        );
    };

    return (
        <AccordionItem
            key={filterKey}
            value={filterKey}
            data-testid={`privilege-zones_details_${filterKey}-accordion-item`}
            className='[&[data-state=open]>div>div>button>svg]:rotate-180 sticky'>
            <div className='w-full flex items-center justify-between border-b border-neutral-3'>
                <div className='flex items-center gap-2'>
                    <Button
                        className='w-6'
                        variant='text'
                        onClick={() => {
                            onOpen((prev) => (prev === filterKey ? '' : filterKey));
                        }}>
                        <FontAwesomeIcon icon={faChevronUp} size='sm' className='font-bold' />
                    </Button>
                    <div className='flex items-center gap'>
                        <SortableHeader
                            title={filterLabels[filterKey]}
                            onSort={() => {
                                setSortOrder((sortOrder) =>
                                    sortOrder === SortOrderAscending ? SortOrderDescending : SortOrderAscending
                                );
                            }}
                            sortOrder={sortOrder}
                            classes={{
                                button: 'font-bold text-base',
                            }}
                        />
                    </div>
                </div>
                <span className='mr-12'>
                    <span className='font-bold'>Total {filterKey.split('_')[0]}: </span>
                    {count}
                </span>
            </div>
            <AccordionContent className='bg-neutral-2 p-0'>
                <div className='border-neutral-5 h-96'>
                    <InfiniteQueryFixedList<AssetGroupTagSelector>
                        itemSize={40}
                        queryResult={rulesQuery}
                        renderRow={Row}
                        renderLoadingRow={LoadingRow}
                    />
                </div>
            </AccordionContent>
        </AccordionItem>
    );
};
