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

import { FC } from 'react';
import { Box, ListItem, ListItemText } from '@mui/material';
import HighlightedText from '../HighlightedText';
import NodeIcon from '../NodeIcon';

type NodeSearchResult = {
    label: string;
    objectId: string;
    kind: string;
};

const SearchResultItem: FC<{
    item: NodeSearchResult;
    index: number;
    highlightedIndex?: number;
    keyword: string;
    getItemProps: (options: any) => any;
}> = ({ item, index, highlightedIndex, keyword, getItemProps }) => {
    return (
        <ListItem
            button
            dense
            selected={highlightedIndex === index}
            key={item.objectId}
            data-testid='explore_search_result-list-item'
            {...getItemProps({ item, index })}>
            <ListItemText
                primary={
                    <Box
                        style={{
                            width: '100%',
                            display: 'flex',
                            alignItems: 'center',
                        }}>
                        <NodeIcon nodeType={item.kind} />
                        <Box
                            style={{
                                flexGrow: 1,
                                marginRight: '1em',
                            }}>
                            <HighlightedText text={item.label || item.objectId} search={keyword} />
                        </Box>
                    </Box>
                }
                primaryTypographyProps={{
                    style: {
                        whiteSpace: 'nowrap',
                        verticalAlign: 'center',
                    },
                }}
            />
        </ListItem>
    );
};

export default SearchResultItem;
