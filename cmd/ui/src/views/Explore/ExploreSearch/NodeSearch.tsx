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

import { useDispatch, useSelector } from 'react-redux';
import ExploreSearchCombobox from '../ExploreSearchCombobox';
import { useEffect, useState } from 'react';
import { AppState } from 'src/store';
import { PRIMARY_SEARCH, SEARCH_TYPE_EXACT, SECONDARY_SEARCH, SearchNodeType } from 'src/ducks/searchbar/types';
import { setSearchValue, startSearchSelected } from 'src/ducks/searchbar/actions';

interface NodeSearchProps {
    labelText: string;
    searchType: typeof PRIMARY_SEARCH | typeof SECONDARY_SEARCH;
}

const NodeSearch = ({ searchType, labelText }: NodeSearchProps) => {
    const dispatch = useDispatch();

    const [inputValue, setInputValue] = useState('');
    const [selectedItem, setSelectedItem] = useState<SearchNodeType | null>(null);

    const searchState = useSelector((state: AppState) => state.search[searchType]);

    useEffect(() => {
        if (searchState.value) {
            setInputValue(searchState.value!.name);
            setSelectedItem(searchState.value);
        }
    }, [searchState]);

    useEffect(() => {
        if (selectedItem) {
            dispatch(setSearchValue(selectedItem, searchType, SEARCH_TYPE_EXACT));
            dispatch(startSearchSelected(searchType));
        }
    }, [selectedItem, searchType, dispatch]);

    const handleInputValueChange = ({ inputValue }: any) => {
        setInputValue(inputValue || '');
    };

    const handleSelectedItemChange = ({ selectedItem }: any) => {
        setSelectedItem(selectedItem);
    };

    return (
        <ExploreSearchCombobox
            inputValue={inputValue}
            onInputValueChange={handleInputValueChange}
            selectedItem={selectedItem}
            onSelectedItemChange={handleSelectedItemChange}
            labelText={labelText}
        />
    );
};

export default NodeSearch;
