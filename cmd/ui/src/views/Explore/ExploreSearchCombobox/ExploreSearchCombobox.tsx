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
import { useCombobox } from 'downshift';
import { useState } from 'react';
import { HighlightedText, NodeIcon } from 'bh-shared-ui';
import { useDebouncedValue } from 'src/hooks/useDebouncedValue';
import { getEmptyResultsText, getKeywordAndTypeValues, SearchResult, useSearch } from 'src/hooks/useSearch';

const ExploreSearchCombobox: React.FC<{
    inputValue: string;
    onInputValueChange: any;
    selectedItem: any;
    onSelectedItemChange: any;
    labelText: string;
    disabled?: boolean;
}> = ({ inputValue, onInputValueChange, selectedItem, onSelectedItemChange, labelText, disabled = false }) => {
    const theme = useTheme();

    const [showInputIcon, setShowInputIcon] = useState(false);

    const debouncedInputValue = useDebouncedValue(inputValue, 150) as string;
    const { keyword, type } = getKeywordAndTypeValues(debouncedInputValue);
    const { data, error, isError, isLoading, isFetching } = useSearch(keyword, type);

    const { isOpen, getMenuProps, getInputProps, getComboboxProps, highlightedIndex, getItemProps, openMenu } =
        useCombobox({
            items: data || [],
            onInputValueChange,
            selectedItem,
            onSelectedItemChange,
            itemToString: (item) => (item ? item.name || item.objectid : ''),
            onStateChange: ({ type, inputValue, selectedItem }) => {
                // remove icon when input is empty
                if (type === useCombobox.stateChangeTypes.InputChange) {
                    if (!inputValue) {
                        setShowInputIcon(false);
                    }
                }

                // show icon when item is selected via click or enter key
                if (type === useCombobox.stateChangeTypes.ItemClick) {
                    setShowInputIcon(true);
                }
                if (type === useCombobox.stateChangeTypes.InputKeyDownEnter) {
                    setShowInputIcon(true);
                }

                // show icon when input is controlled via props
                if (type === useCombobox.stateChangeTypes.ControlledPropUpdatedSelectedItem) {
                    setShowInputIcon(true);
                }
            },
        });

    const disabledText: string = getEmptyResultsText(
        isLoading,
        isFetching,
        isError,
        error,
        debouncedInputValue,
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
                    startAdornment: showInputIcon && selectedItem?.type && <NodeIcon nodeType={selectedItem?.type} />,
                }}
                {...getInputProps({ onFocus: openMenu })}
                data-testid='explore_search_input-search'
            />
            <div
                style={{
                    position: 'absolute',
                    marginTop: '1rem',
                    minWidth: '360px',
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
                                    <ListItem
                                        button
                                        dense
                                        selected={highlightedIndex === index}
                                        key={item.objectid}
                                        data-testid='explore_search_result-list-item'
                                        {...getItemProps({ item, index })}>
                                        <ListItemText
                                            primary={
                                                <div
                                                    style={{
                                                        width: '100%',
                                                        display: 'flex',
                                                        alignItems: 'center',
                                                    }}>
                                                    <NodeIcon nodeType={item.type} />
                                                    <div
                                                        style={{
                                                            flexGrow: 1,
                                                            marginRight: '1em',
                                                        }}>
                                                        <HighlightedText
                                                            text={item.name || item.objectid}
                                                            search={keyword}
                                                        />
                                                    </div>
                                                </div>
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
                            })
                        )}
                    </List>
                </Paper>
            </div>
        </div>
    );
};

export default ExploreSearchCombobox;
