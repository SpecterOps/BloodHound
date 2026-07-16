// Copyright 2026 Specter Ops, Inc.
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
import { faChevronDown } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Accordion, AccordionContent, AccordionHeader, AccordionItem, Tooltip } from 'doodle-ui';
import { NodeKindInfo, NodeKindInfoItem, RelationshipKindInfo, RelationshipKindInfoItem } from 'js-client-library';
import { useCallback } from 'react';
import { useExploreParams } from '../../hooks/useExploreParams';
import MarkdownContent from '../MarkdownContent';

type KindInfoItemProps = {
    item: NodeKindInfoItem | RelationshipKindInfoItem;
};

function KindInfoItem({ item }: KindInfoItemProps) {
    return (
        <AccordionItem
            data-testid='entity-info_kind-info-item'
            value={item.name}
            className='h-fit align-middle border-t border-neutral-400 dark:border-neutral-700'>
            <AccordionHeader className='text-sm p-0 pt-2 pb-3 my-2 -mx-4 px-4 w-full h-full dark:hover:bg-neutral-700 hover:bg-neutral-200 hover:no-underline'>
                <FontAwesomeIcon icon={faChevronDown} size='sm' />
                <Tooltip tooltip={item.title}>
                    <span className='ml-3 truncate'>{item.title}</span>
                </Tooltip>
            </AccordionHeader>
            <AccordionContent>
                <MarkdownContent markdown={item.markdown?.content} />
            </AccordionContent>
        </AccordionItem>
    );
}

type KindInfoItemsProps = {
    items: NodeKindInfo | RelationshipKindInfo | undefined;
};

const sortItems = (items: NodeKindInfo | RelationshipKindInfo) => {
    return [...items].sort((a, b) => {
        // Extract the appropriate ID based on the item type
        const aId = 'node_kind_id' in a ? a.node_kind_id : a.relationship_kind_id;
        const bId = 'node_kind_id' in b ? b.node_kind_id : b.relationship_kind_id;

        return a.position - b.position || a.title.localeCompare(b.title) || aId - bId;
    });
};

export function KindInfoItems({ items }: KindInfoItemsProps) {
    const { setExploreParams, expandedPanelSections } = useExploreParams();

    const handleChange = useCallback(
        (values: string[]) => {
            setExploreParams({ expandedPanelSections: values });
        },
        [setExploreParams]
    );

    if (!items) return null;

    return (
        <Accordion
            data-testid='entity-info_kind-info-items'
            type='multiple'
            value={expandedPanelSections ? expandedPanelSections : undefined}
            className='p-0 pt-2 bg-neutral-2'
            onValueChange={handleChange}>
            {sortItems(items).map((item) => {
                return <KindInfoItem key={item.name} item={item} />;
            })}
        </Accordion>
    );
}
