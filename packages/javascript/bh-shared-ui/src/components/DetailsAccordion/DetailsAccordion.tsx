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

import { Accordion, AccordionContent, AccordionHeader, AccordionItem } from '@bloodhoundenterprise/doodleui';
import { faAngleDown } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import clsx from 'clsx';

import React, { ComponentType } from 'react';

export type DetailsAccordionProps<T extends Record<string, unknown>> = {
    /** If `true`, applies an accent style to item headers  */
    accent?: boolean;

    /** Component to render the content area of each item */
    Content: ComponentType<T>;

    /** Optional component to render when `items` is `undefined` */
    Empty?: ComponentType;

    /** Function to provide a stable key for each row (defaults to index) */
    getKey?: (item: T, index: number) => React.Key;

    /** Predicate to determine if an item is disabled (default: always `false`) */
    itemDisabled?: (item: T) => boolean;

    /** Component to render the header of each item */
    Header: ComponentType<T>;

    /** A single item or array of items to render */
    items?: null | T | T[];

    /** Index of the item that should start open (`undefined` = none) */
    openIndex?: number;
};

/**
 * A generic accordion component for displaying a list of items, where each
 * item has a header and expandable content.
 *
 * @typeParam T - Shape of a rendered item. Extends `Record<string, unknown>`.
 *
 * @remarks
 * - Renders each item inside a single-select accordion panel.
 * - If `items` is `undefined` and an `Empty` component is provided, the `Empty` component will render instead.
 * - Each item is passed as props to the `Header` and `Content` components.
 * - Supports disabling items, custom keys, and optionally applying an accent style.
 *
 * @example
 * ```tsx
 * <DetailsAccordion
 *   Header={({ title }) => <div>{title}</div>}
 *   Content={({ description }) => <p>{description}</p>}
 *   items={[
 *     { title: "First", description: "First item description" },
 *     { title: "Second", description: "Second item description" },
 *   ]}
 *   openIndex={0}
 * />
 * ```
 */
export const DetailsAccordion = <T extends Record<string, unknown>>({
    accent = false,
    Content,
    Empty,
    getKey,
    items,
    itemDisabled = () => false,
    Header,
    openIndex,
}: DetailsAccordionProps<T>) => {
    if (items === undefined && Empty) {
        return <Empty />;
    }

    const itemArray = Array.isArray(items) ? items : [items];

    const defaultValue =
        typeof openIndex === 'number' && openIndex >= 0 && openIndex < itemArray.length ? String(openIndex) : undefined;

    return (
        <Accordion collapsible defaultValue={defaultValue} type='single'>
            {itemArray.map((item, idx) => {
                if (item === undefined || item === null) {
                    return null;
                }

                const key = getKey ? getKey(item, idx) : idx;
                const isDisabled = itemDisabled(item);

                return (
                    <AccordionItem
                        className='bg-neutral-light-2 dark:bg-neutral-dark-2 border-t dark:border-neutral-dark-4 first:border-none'
                        disabled={isDisabled}
                        key={key}
                        value={String(idx)}>
                        <AccordionHeader
                            className={clsx(
                                'h-16',
                                !isDisabled && 'cursor-pointer',
                                isDisabled && 'hover:no-underline',
                                !accent && 'ml-2',
                                accent && 'bg-[#e0e0e0] dark:bg-[#202020] border-l-8 border-primary'
                            )}>
                            <FontAwesomeIcon
                                aria-hidden
                                className={clsx(isDisabled && 'opacity-0')}
                                icon={faAngleDown}
                            />
                            <Header {...item} />
                        </AccordionHeader>

                        <AccordionContent className='p-0 text-base border-t dark:border-neutral-dark-4'>
                            <Content {...item} />
                        </AccordionContent>
                    </AccordionItem>
                );
            })}
        </Accordion>
    );
};
