import { Accordion, AccordionContent, AccordionHeader, AccordionItem } from '@bloodhoundenterprise/doodleui';
import { faAngleDown } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import clsx from 'clsx';
import { stubFalse } from 'lodash';

import React, { ComponentType } from 'react';

type DetailsAccordionProps<T extends Record<string, unknown>> = {
    /** If `true`, applies an accent style to the selected item’s header  */
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
    items?: T | T[];

    /** Index of the item that should start open (`null` = none) */
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
    itemDisabled = stubFalse,
    Header,
    openIndex,
}: DetailsAccordionProps<T>) => {
    if (items === undefined && Empty) {
        return <Empty />;
    }

    const itemArray = Array.isArray(items) ? items : [items];

    return (
        <Accordion collapsible defaultValue={String(openIndex)} type='single'>
            {itemArray.map((item, idx) => {
                if (item === undefined) {
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
                                accent && 'bg-[#e0e0e0] dark:bg-[#202020] border-l-8 border-primary'
                            )}>
                            <FontAwesomeIcon icon={faAngleDown} className={clsx(isDisabled && 'opacity-0')} />
                            <Header {...item} />
                        </AccordionHeader>

                        <AccordionContent className='pr-6 text-base justify-between border-t dark:border-neutral-dark-4'>
                            <Content {...item} />
                        </AccordionContent>
                    </AccordionItem>
                );
            })}
        </Accordion>
    );
};
