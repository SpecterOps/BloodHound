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
import { faCaretRight, faChevronUp } from '@fortawesome/free-solid-svg-icons';
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
import { SelectedHighlight } from '../Details/SelectedHighlight';

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

const SelectedCaretRight = () => {
    return (
        <div className='absolute right-4 bottom-2 text-sm'>
            <FontAwesomeIcon icon={faCaretRight} />
        </div>
    );
};

export const RulesAccordion: React.FC = () => {
    const [openAccordion, setOpenAccordion] = useState<RuleSection | ''>(CustomRulesKey);
    const selectedTag = useSelectedTagPathParams();
    const { ruleId, tagDetailsLink, tagId, isZonePage } = usePZPathParams();
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
                className={cn('border-y border-neutral-3 relative', {
                    'bg-neutral-4': !ruleId,
                })}>
                {selectedTag.id && <SelectedHighlight itemId={selectedTag.id} type='tag' />}
                <Button
                    variant='text'
                    className='w-full block text-left'
                    onClick={() => navigate(tagDetailsLink(tagId))}>
                    <span className='pl-6 text-base text-contrast ml-2 truncate'>All Rules</span>
                </Button>
                {!ruleId && <SelectedCaretRight />}
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
                    isOpen={openAccordion === CustomRulesKey}
                    onOpen={setOpenAccordion}
                />
                {isZonePage && (
                    <RuleAccordionItem
                        section={DefaultRulesKey}
                        count={selectedTag.counts[DefaultRulesKey]}
                        isOpen={openAccordion === DefaultRulesKey}
                        onOpen={setOpenAccordion}
                    />
                )}
                <RuleAccordionItem
                    section={DisabledRulesKey}
                    count={selectedTag.counts[DisabledRulesKey]}
                    isOpen={openAccordion === DisabledRulesKey}
                    onOpen={setOpenAccordion}
                />
            </Accordion>
        </div>
    );
};

interface RuleAccordionItemProps {
    section: RuleSection;
    count: number;
    isOpen: boolean;
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

const RuleAccordionItem: React.FC<RuleAccordionItemProps> = ({ section: filterKey, count, isOpen, onOpen }) => {
    const [sortOrder, setSortOrder] = useState<SortOrder>(SortOrderAscending);

    const navigate = useAppNavigate();

    const { ruleId, tagId, ruleDetailsLink } = usePZPathParams();

    const environments = useEnvironmentIdList([{ path: `/${privilegeZonesPath}/*`, caseSensitive: false, end: false }]);

    const rulesQuery = useRulesInfiniteQuery(tagId, { sortOrder, environments, ...filters[filterKey] }, isOpen);

    const isRuleSelected = (id: string) => ruleId === id;

    const isAccordionDisabled = count === 0;

    const handleClick = (id: number) => navigate(ruleDetailsLink(tagId, id));

    const Row: InfiniteQueryFixedListProps<AssetGroupTagSelector>['renderRow'] = (item, index, style) => {
        return (
            <Tooltip
                tooltip={<span className='text-contrast'>{item.name}</span>}
                contentProps={{ className: 'bg-neutral-3' }}>
                <div
                    key={item.id}
                    role='listitem'
                    className={cn('border-y border-neutral-3 relative', {
                        'bg-neutral-4': isRuleSelected(item.id.toString()),
                    })}
                    style={style}>
                    <SelectedHighlight itemId={item.id} type='rule' />
                    <Button
                        variant='text'
                        className='w-full block text-left truncate'
                        title={`Name: ${item.name}`}
                        onClick={() => handleClick(item.id)}>
                        <span className='pl-6 text-base text-contrast ml-2'>{item.name}</span>
                    </Button>
                    {isRuleSelected(item.id.toString()) && <SelectedCaretRight />}
                </div>
            </Tooltip>
        );
    };

    return (
        <AccordionItem
            key={filterKey}
            value={filterKey}
            data-testid={`privilege-zones_details_${filterKey}-accordion-item`}
            className='[&[data-state=open]>div>div>button>svg]:rotate-180 sticky'>
            <div className='w-full flex items-center justify-between border-b border-neutral-3'>
                <div className='w-full flex items-center gap-2'>
                    <Button
                        className='w-6 max-xl:px-2 max-lg:px-6'
                        variant='text'
                        disabled={isAccordionDisabled}
                        data-testid={`privilege-zones_details_${filterKey}-accordion_open-toggle-button`}
                        onClick={() => {
                            onOpen((prev) => (prev === filterKey ? '' : filterKey));
                        }}>
                        <FontAwesomeIcon icon={faChevronUp} size='sm' className='font-bold' />
                    </Button>
                    <div className='flex-1 items-center gap'>
                        <SortableHeader
                            title={filterLabels[filterKey]}
                            disable={isAccordionDisabled}
                            onSort={() => {
                                setSortOrder((sortOrder) =>
                                    sortOrder === SortOrderAscending ? SortOrderDescending : SortOrderAscending
                                );
                            }}
                            sortOrder={sortOrder}
                            classes={{
                                container: cn({ 'pointer-events-none cursor-default': !isOpen }),
                                button: cn('font-bold text-base', {
                                    '[&>svg]:hidden': !isOpen || isAccordionDisabled,
                                    'opacity-50': isAccordionDisabled,
                                }),
                            }}
                        />
                    </div>
                </div>
                <span className='pr-12 max-xl:pr-4 max-lg:pr-12 flex-none'>
                    <span className='font-bold'>
                        Total <span className='capitalize'>{filterKey.split('_')[0]}</span>:{' '}
                    </span>
                    {count.toLocaleString()}
                </span>
            </div>
            <AccordionContent className='bg-neutral-2 p-0'>
                <div className='border-neutral-5 h-80 lg:h-80 xl:h-[28rem] 2xl:h-[36rem]'>
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
