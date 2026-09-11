// Copyright 2023 Specter Ops, Inc.
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

import { ListItem, ListItemText } from '@mui/material';
import { Typography } from 'doodle-ui';
import { FC } from 'react';
import { cn } from '../../utils';
import HighlightedText from '../HighlightedText';
import NodeIcon from '../NodeIcon';

export type NodeSearchResult = {
    label: string;
    objectId: string;
    kind: string;
    id?: string;
    distinguishedName?: string;
};

const SearchResultItem: FC<{
    item: NodeSearchResult;
    index: number;
    getItemProps: (options: any) => any;
    highlightedIndex?: number;
    style?: React.CSSProperties;
    keyword?: string;
}> = ({ style, item, index, highlightedIndex, keyword, getItemProps }) => {
    return (
        <ListItem
            dense
            style={style}
            className={cn(
                'group cursor-pointer hover:bg-secondary hover:text-common-white hover:dark:bg-secondary-variant-2 hover:dark:text-common-dark focus:bg-secondary focus:text-common-white focus:dark:bg-secondary-variant-2 focus:dark:text-common-dark focus-visible:bg-secondary focus-visible:text-common-white focus-visible:dark:bg-secondary-variant-2 focus-visible:dark:text-common-dark',
                {
                    'bg-secondary text-common-white dark:bg-secondary-variant-2 dark:text-common-dark':
                        highlightedIndex === index,
                }
            )}
            key={item.objectId}
            data-testid='explore_search_result-list-item'
            tabIndex={0}
            {...getItemProps({ item, index })}>
            <ListItemText
                disableTypography
                primary={
                    <div className='flex items-start whitespace-nowrap'>
                        <NodeIcon
                            nodeType={item.kind}
                            className={cn(
                                'group-hover:text-inherit group-focus:text-inherit group-focus-visible:text-inherit',
                                {
                                    'text-inherit': highlightedIndex === index,
                                }
                            )}
                        />
                        <div className='flex flex-col'>
                            <HighlightedText text={item.label || item.objectId} search={keyword} />
                            {item.distinguishedName && (
                                <Typography
                                    variant='caption'
                                    className={cn(
                                        // TODO: Tokenize when available
                                        'text-[#505050] dark:text-[#CDCDCD] group-hover:text-inherit group-focus:text-inherit group-focus-visible:text-inherit',
                                        {
                                            'text-inherit': highlightedIndex === index,
                                        }
                                    )}>
                                    {item.distinguishedName}
                                </Typography>
                            )}
                        </div>
                    </div>
                }
            />
        </ListItem>
    );
};

export default SearchResultItem;
