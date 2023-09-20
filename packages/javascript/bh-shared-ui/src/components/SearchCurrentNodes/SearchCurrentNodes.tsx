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

import { Box, List, ListItem, Paper, SxProps, TextField } from '@mui/material';
import { useCombobox } from 'downshift';
import { FC, useEffect, useRef, useState } from 'react';
import SearchResultItem from '../SearchResultItem';
import { FlatNode, GraphNodes } from './types';
import { useOnClickOutside } from '../../hooks';
import { FixedSizeList } from 'react-window';

export const PLACEHOLDER_TEXT = 'Search Current Results';
export const NO_RESULTS_TEXT = 'No result found in current results';

const LIST_ITEM_HEIGHT = 38;
const MAX_CONTAINER_HEIGHT = 350;

const SearchCurrentNodes: FC<{
    sx?: SxProps;
    currentNodes: GraphNodes;
    onSelect: (node: FlatNode) => void;
    onClose?: () => void;
}> = ({ sx, currentNodes, onSelect, onClose }) => {
    const containerRef = useRef<HTMLDivElement>(null);
    const inputRef = useRef<HTMLInputElement>(null);

    const [flatNodeList, setFlatNodeList] = useState<FlatNode[]>([]);
    const [items, setItems] = useState<FlatNode[]>([]);
    const [selectedNode, setSelectedNode] = useState<FlatNode | null | undefined>(null);
    const [virtualizationHeight, setVirtualizationHeight] = useState<number>(0);

    // Node data is a lot easier to work with in the combobox if we transform to an array of flat objects
    useEffect(() => {
        const flatNodeList: FlatNode[] = Object.entries(currentNodes).map(([key, value]) => {
            return { id: key, ...value };
        });
        setFlatNodeList(flatNodeList);
    }, [currentNodes]);

    useEffect(() => inputRef.current?.focus(), []);

    useEffect(() => {
        if (selectedNode) onSelect(selectedNode);
    }, [selectedNode, onSelect]);

    // Since we are using a virtualized results container, we need to calculate the height for shorter
    // lists to avoid whitespace
    useEffect(() => {
        const resultsHeight = LIST_ITEM_HEIGHT * items.length;
        if (resultsHeight > MAX_CONTAINER_HEIGHT) {
            setVirtualizationHeight(MAX_CONTAINER_HEIGHT - 10);
        } else {
            setVirtualizationHeight(resultsHeight);
        }
    }, [items]);

    useOnClickOutside(containerRef, () => onClose && onClose());

    const { getInputProps, getMenuProps, getComboboxProps, getItemProps, inputValue } = useCombobox({
        items,
        onInputValueChange: ({ inputValue }) => {
            const filteredNodes = flatNodeList.filter((node) => {
                const label = node.label.toLowerCase();
                const objectId = node.objectId.toLowerCase();
                const lowercaseInputValue = inputValue?.toLowerCase() || '';

                if (inputValue === '') return false;
                return label.includes(lowercaseInputValue) || objectId.includes(lowercaseInputValue);
            });
            setItems(filteredNodes);
        },
        stateReducer: (_state, actionAndChanges) => {
            const { changes, type } = actionAndChanges;
            switch (type) {
                case useCombobox.stateChangeTypes.ItemClick:
                    if (changes.selectedItem) setSelectedNode(changes.selectedItem);
                    return { ...changes, inputValue: '' };
                default:
                    return changes;
            }
        },
    });

    const Row = ({ index, style }: any) => {
        return (
            <Box style={style} overflow={'hidden'}>
                <SearchResultItem
                    item={items[index]}
                    index={index}
                    key={index}
                    highlightedIndex={0}
                    keyword={inputValue}
                    getItemProps={getItemProps}
                />
            </Box>
        );
    };

    return (
        <div ref={containerRef}>
            <Box component={Paper} {...sx} {...getComboboxProps()}>
                <Box overflow={'auto'} maxHeight={MAX_CONTAINER_HEIGHT} marginBottom={items.length === 0 ? 0 : 1}>
                    <List
                        data-testid={'current-results-list'}
                        dense
                        {...getMenuProps({
                            hidden: items.length === 0 && !inputValue,
                            style: { paddingTop: 0 },
                        })}>
                        {
                            <FixedSizeList
                                height={virtualizationHeight}
                                width={'100%'}
                                itemSize={LIST_ITEM_HEIGHT}
                                itemCount={items.length}>
                                {Row}
                            </FixedSizeList>
                        }
                        {items.length === 0 && inputValue && (
                            <ListItem disabled sx={{ fontSize: 14 }}>
                                {NO_RESULTS_TEXT}
                            </ListItem>
                        )}
                    </List>
                </Box>
                <TextField
                    inputRef={inputRef}
                    placeholder={PLACEHOLDER_TEXT}
                    variant='outlined'
                    size='small'
                    fullWidth
                    {...getInputProps()}
                    InputProps={{ sx: { fontSize: 14 } }}
                />
            </Box>
        </div>
    );
};

export default SearchCurrentNodes;
