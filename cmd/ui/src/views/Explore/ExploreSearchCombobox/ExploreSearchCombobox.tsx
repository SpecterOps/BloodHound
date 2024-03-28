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

import { List, ListItem, ListItemText, Paper, TextField, useTheme } from '@mui/material';
import {
    NodeIcon,
    SearchValue,
    SearchResult,
    SearchResultItem,
    getEmptyResultsText,
    getKeywordAndTypeValues,
    useSearch,
} from 'bh-shared-ui';
import { useCombobox } from 'downshift';

const ExploreSearchCombobox: React.FC<{
    labelText: string;
    inputValue: string;
    selectedItem: SearchValue | null;
    handleNodeEdited: (edit: string) => any;
    handleNodeSelected: (selection: SearchValue) => any;
    disabled?: boolean;
}> = ({ labelText, inputValue, selectedItem, handleNodeEdited, handleNodeSelected, disabled = false }) => {
    const theme = useTheme();

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

    return (
        <div {...getComboboxProps()} style={{ position: 'relative' }}>
            <TextField
                placeholder={labelText}
                variant='outlined'
                size='small'
                fullWidth
                disabled={disabled}
                inputProps={{
                    'aria-label': labelText,
                }}
                InputProps={{
                    style: {
                        backgroundColor: disabled ? theme.palette.action.disabled : 'inherit',
                        fontSize: theme.typography.pxToRem(14),
                    },
                    startAdornment: selectedItem?.type && <NodeIcon nodeType={selectedItem?.type} />,
                }}
                {...getInputProps({
                    onFocus: openMenu,
                    refKey: 'inputRef',
                    onChange: (e) => {
                        handleNodeEdited(e.currentTarget.value);
                    },
                })}
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
                            data!.map((item: SearchResult, index: any) => {
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
