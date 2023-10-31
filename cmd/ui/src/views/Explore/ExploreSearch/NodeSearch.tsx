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
import { AppState } from 'src/store';
import { PRIMARY_SEARCH, SECONDARY_SEARCH } from 'src/ducks/searchbar/types';
import { startSearchSelected } from 'src/ducks/searchbar/actions';
import { useEffect } from 'react';

interface NodeSearchProps {
    labelText: string;
    searchType: typeof PRIMARY_SEARCH | typeof SECONDARY_SEARCH;
}

const NodeSearch = ({ searchType, labelText }: NodeSearchProps) => {
    const dispatch = useDispatch();
    const { primary } = useSelector((state: AppState) => state.search);

    useEffect(() => {
        if (primary.value) {
            dispatch(startSearchSelected(PRIMARY_SEARCH));
        }
    }, [primary, dispatch]);

    return <ExploreSearchCombobox labelText={labelText} searchType={searchType} />;
};

export default NodeSearch;
