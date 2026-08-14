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

import { faCaretRight, faChevronUp } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Accordion, AccordionContent, AccordionItem, IconButton, Skeleton, TextButton, Tooltip } from 'doodle-ui';
import { AssetGroupTagSelector, CustomRulesKey, DefaultRulesKey, DisabledRulesKey, RulesKey } from 'js-client-library';
import { useEffect, useRef, useState } from 'react';
import { FixedSizeList } from 'react-window';
import { SortableHeader } from '../../../components/ColumnHeaders';
import { optionStyles } from '../../../components/DropdownSelector/constants';
import { InfiniteQueryFixedList, InfiniteQueryFixedListProps } from '../../../components/InfiniteQueryFixedList';
import { useRuleInfo, useRulesInfiniteQuery } from '../../../hooks/useAssetGroupTags';
import { useEnvironmentIdList } from '../../../hooks/useEnvironmentIdList';
import { usePZPathParams } from '../../../hooks/usePZParams/usePZPathParams';
import { useSelectedTagPathParams } from '../../../hooks/useSelectedTag';
import { ENVIRONMENT_AGGREGATION_SUPPORTED_ROUTES } from '../../../routes';
import { SortOrder, SortOrderAscending, SortOrderDescending } from '../../../types';
import { cn, useAppNavigate } from '../../../utils';
import { RuleTabValue, TagTabValue } from '../utils';
import { useSelectedDetailsTabsContext } from './SelectedDetailsTabs/SelectedDetailsTabsContext';
import { SelectedHighlight } from './SelectedHighlight';

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
    const { setSelectedDetailsTab } = useSelectedDetailsTabsContext();
    const { data: selectedRule } = useRuleInfo(tagId?.toString() ?? '', ruleId ?? '');

    useEffect(() => {
        if (!selectedRule) return;

        if (selectedRule?.disabled_at) {
            setOpenAccordion(DisabledRulesKey);
        } else if (selectedRule?.is_default) {
            setOpenAccordion(DefaultRulesKey);
        } else {
            setOpenAccordion(CustomRulesKey);
        }
    }, [selectedRule]);

    if (!selectedTag.counts) return null;

    return (
        <div>
            <div className='flex justify-between items-center pl-4 py-2 pr-12 border-b border-neutral-3'>
                <span className='text-lg font-bold'>Rules</span>
                <span>
                    <span className='font-bold'>Total Rules:</span> {selectedTag.counts[RulesKey].toLocaleString()}
                </span>
            </div>
            <div
                className={cn('border-b border-neutral-3 relative', {
                    'bg-neutral-4': !ruleId,
                })}>
                {selectedTag.id && <SelectedHighlight itemId={selectedTag.id} type='tag' />}
                <TextButton
                    className={cn(optionStyles, 'block text-left h-10 px-1')}
                    onClick={() => {
                        setSelectedDetailsTab(TagTabValue);
                        navigate(tagDetailsLink(tagId));
                    }}>
                    <span className='pl-6 text-base ml-2'>All Rules</span>
                </TextButton>
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
        className='border-b border-neutral-3 relative w-full p-2'>
        <Skeleton className='h-full' />
    </div>
);

const RuleAccordionItem: React.FC<RuleAccordionItemProps> = ({ section: filterKey, count, isOpen, onOpen }) => {
    const listRef = useRef<FixedSizeList<AssetGroupTagSelector[]>>(null);
    const [sortOrder, setSortOrder] = useState<SortOrder>(SortOrderAscending);

    const navigate = useAppNavigate();

    const { ruleId, tagId, ruleDetailsLink } = usePZPathParams();

    const { setSelectedDetailsTab } = useSelectedDetailsTabsContext();

    const environments = useEnvironmentIdList(ENVIRONMENT_AGGREGATION_SUPPORTED_ROUTES, false);

    const rulesQuery = useRulesInfiniteQuery(tagId, { sortOrder, environments, ...filters[filterKey] }, isOpen);

    const isRuleSelected = (id: string) => ruleId === id;
    const isAccordionDisabled = count === 0;

    const handleClick = (id: number) => {
        setSelectedDetailsTab(RuleTabValue);
        navigate(ruleDetailsLink(tagId, id));
    };

    useEffect(() => {
        const ruleInAccordion = isOpen && ruleId;

        if (!ruleInAccordion) return;

        const { fetchNextPage, hasNextPage, isFetchingNextPage } = rulesQuery;
        const ruleIdNumber = Number(ruleId);

        if (!Number.isFinite(ruleIdNumber)) return;

        const allItems = rulesQuery?.data?.pages.flatMap((page) => page.items);
        const selectedItemIndex = allItems?.findIndex((rule) => rule.id === ruleIdNumber);

        if (typeof selectedItemIndex === 'number' && selectedItemIndex > -1) {
            listRef.current?.scrollToItem(selectedItemIndex, 'smart');
        }

        if (selectedItemIndex === -1 && hasNextPage && !isFetchingNextPage) {
            fetchNextPage();
        }
    }, [ruleId, isOpen, rulesQuery]);

    const Row: InfiniteQueryFixedListProps<AssetGroupTagSelector>['renderRow'] = (item, index, style) => {
        const isSelected = isRuleSelected(item.id.toString());

        return (
            <Tooltip
                tooltip={<span className='text-contrast'>{item.name}</span>}
                contentProps={{ className: 'bg-neutral-3' }}>
                <div
                    key={item.id}
                    role='listitem'
                    className={cn('border-b border-neutral-3 relative', {
                        'bg-neutral-4': isSelected,
                    })}
                    style={style}>
                    <SelectedHighlight itemId={item.id} type='rule' />
                    <TextButton className={cn(optionStyles, 'px-1')} onClick={() => handleClick(item.id)}>
                        <span className='pl-8 text-base ml-3.5'>{item.name}</span>
                    </TextButton>
                    {isSelected && <SelectedCaretRight />}
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
                <div className='w-full flex items-center h-10'>
                    <IconButton
                        className='mx-2 rounded-none'
                        aria-label={isOpen ? 'Collapse' : 'Expand'}
                        disabled={isAccordionDisabled}
                        data-testid={`privilege-zones_details_${filterKey}-accordion_open-toggle-button`}
                        onClick={() => {
                            onOpen((prev) => (prev === filterKey ? '' : filterKey));
                        }}>
                        <FontAwesomeIcon icon={faChevronUp} size='sm' className='font-bold' />
                    </IconButton>
                    <div className='flex items-center gap-2'>
                        <SortableHeader
                            title={filterLabels[filterKey]}
                            disable={!isOpen || isAccordionDisabled}
                            onSort={() => {
                                setSortOrder((sortOrder) =>
                                    sortOrder === SortOrderAscending ? SortOrderDescending : SortOrderAscending
                                );
                            }}
                            sortOrder={sortOrder}
                            classes={{
                                container: cn({ 'pointer-events-none cursor-default': !isOpen }),
                                button: cn('font-bold text-base rounded-none', {
                                    '[&>svg]:hidden': !isOpen || isAccordionDisabled,
                                    'opacity-50': isAccordionDisabled,
                                    'disabled:!text-text-main disabled:!opacity-100 disabled:dark:!text-common-white disabled:dark:!opacity-100':
                                        !isOpen && !isAccordionDisabled,
                                }),
                            }}
                        />
                    </div>
                </div>
                <span className='mr-12 max-xl:pr-4 max-lg:pr-12 flex-none'>
                    <span className='font-bold'>
                        Total <span className='capitalize'>{filterKey.split('_')[0]}</span>:{' '}
                    </span>
                    {count.toLocaleString()}
                </span>
            </div>
            <AccordionContent className='bg-neutral-2 p-0'>
                <div className='border-neutral-5'>
                    <InfiniteQueryFixedList<AssetGroupTagSelector>
                        listRef={listRef}
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
