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

import { List, ListItem, ListItemText, Paper, TextField, TextFieldVariants } from '@mui/material';
import { useCombobox } from 'downshift';
import { useRef } from 'react';
import {
    SearchResult,
    getEmptyResultsText,
    getKeywordAndTypeValues,
    useKeybindings,
    useSearch,
    useTheme,
} from '../../hooks';
import { SearchValue } from '../../views/Explore/ExploreSearch/types';
import NodeIcon from '../NodeIcon';
import SearchResultItem from '../SearchResultItem';

const ExploreSearchCombobox: React.FC<{
    labelText: string;
    inputValue: string;
    autoFocus?: boolean;
    selectedItem: SearchValue | null;
    handleNodeEdited: (edit: string) => any;
    handleNodeSelected: (selection: SearchValue) => any;
    disabled?: boolean;
    variant?: TextFieldVariants;
}> = ({
    labelText,
    inputValue,
    selectedItem,
    handleNodeEdited,
    handleNodeSelected,
    autoFocus,
    disabled = false,
    variant = 'outlined',
}) => {
    const theme = useTheme();
    const searchNodesRef = useRef<HTMLInputElement>();

    const { keyword, type } = getKeywordAndTypeValues(inputValue);
    const { data, error, isError, isLoading, isFetching } = useSearch(keyword, type);

    const { isOpen, getMenuProps, getInputProps, getComboboxProps, highlightedIndex, getItemProps, openMenu } =
        useCombobox({
            items: data || [],
            inputValue,
            selectedItem,
            onSelectedItemChange: ({ selectedItem }) => {
                if (selectedItem) {
                    handleNodeSelected(selectedItem);
                }
            },
            itemToString: (item) => item?.name || item?.objectid || '',
        });

    const disabledText: string = getEmptyResultsText(
        isLoading,
        isFetching,
        isError,
        error,
        inputValue,
        type,
        keyword,
        data
    );

    const downshiftInputProps = {
        ...getInputProps({
            onFocus: openMenu,
            refKey: 'inputRef',
            onChange: (e) => {
                handleNodeEdited(e.currentTarget.value);
            },
        }),
    };

    useKeybindings({
        Slash: () => {
            searchNodesRef.current?.focus();
        },
    });

    return (
        <div {...getComboboxProps()} style={{ position: 'relative' }}>
            <TextField
                placeholder={labelText}
                variant={variant}
                size='small'
                fullWidth
                disabled={disabled}
                inputProps={{
                    'aria-label': labelText,
                }}
                InputProps={{
                    style: {
                        backgroundColor: disabled ? theme.neutral.tertiary : 'inherit',
                        fontSize: '0.875rem',
                    },
                    autoFocus,
                    startAdornment: selectedItem?.type && <NodeIcon nodeType={selectedItem?.type} />,
                }}
                {...downshiftInputProps}
                inputRef={(node) => {
                    downshiftInputProps.inputRef(node);
                    searchNodesRef.current = node;
                }}
                data-testid='explore_search_input-search'
            />
            <div
                style={{
                    position: 'absolute',
                    marginTop: '1rem',
                    zIndex: 1300,
                }}>
                <Paper style={{ display: isOpen ? 'inherit' : 'none' }}>
                    <List
                        {...getMenuProps()}
                        dense
                        style={{
                            width: '100%',
                        }}
                        data-testid='explore_search_result-list'>
                        {disabledText ? (
                            <ListItem disabled dense>
                                <ListItemText primary={disabledText} />
                            </ListItem>
                        ) : (
                            data?.map((item: SearchResult, index: any) => {
                                return (
                                    <SearchResultItem
                                        item={{
                                            label: item.name,
                                            objectId: item.objectid,
                                            kind: item.type,
                                        }}
                                        index={index}
                                        key={index}
                                        highlightedIndex={highlightedIndex}
                                        keyword={keyword}
                                        getItemProps={getItemProps}
                                    />
                                );
                            })
                        )}
                    </List>
                </Paper>
            </div>
        </div>
    );
};

export default ExploreSearchCombobox;
