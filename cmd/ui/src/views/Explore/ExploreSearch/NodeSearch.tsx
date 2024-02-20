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

import ExploreSearchCombobox from '../ExploreSearchCombobox';
import { SearchNodeType, SourceNodeEditedAction, SourceNodeSelectedAction } from 'src/ducks/searchbar/types';
import { useAppDispatch, useAppSelector } from 'src/store';
import { sourceNodeEdited, sourceNodeSelected } from 'src/ducks/searchbar/actions';

const NodeSearch = () => {
    const dispatch = useAppDispatch();

    const primary = useAppSelector((state) => state.search.primary);
    const { searchTerm, value: selectedItem } = primary;

    const handleNodeEdited = (edit: string): SourceNodeEditedAction => dispatch(sourceNodeEdited(edit));
    const handleNodeSelected = (selected: SearchNodeType): SourceNodeSelectedAction =>
        dispatch(sourceNodeSelected(selected));

    return (
        <ExploreSearchCombobox
            labelText={'Search Nodes'}
            inputValue={searchTerm}
            selectedItem={selectedItem}
            handleNodeEdited={handleNodeEdited}
            handleNodeSelected={handleNodeSelected}
        />
    );
};

export default NodeSearch;
