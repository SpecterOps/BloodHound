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

import { Autocomplete, AutocompleteRenderInputParams, TextField } from '@mui/material';
import { FC, HTMLAttributes, ReactNode, SyntheticEvent, useState } from 'react';
import AutocompleteOption from './AutocompleteOption';
import { AssetGroupChangelog, AssetGroupChangelogEntry, ChangelogAction } from './types';
import { AssetGroup } from 'js-client-library';
import { getEmptyResultsText, getKeywordAndTypeValues, useDebouncedValue, useSearch } from '../../hooks';

export const AUTOCOMPLETE_PLACEHOLDER = 'Add or Remove Members';

const AssetGroupAutocomplete: FC<{
    assetGroup: AssetGroup;
    changelog: AssetGroupChangelog;
    onChange: (event: any, value: AssetGroupChangelogEntry) => void;
}> = ({ assetGroup, changelog, onChange }) => {
    const [searchInput, setSearchInput] = useState('');
    const debouncedInputValue = useDebouncedValue(searchInput, 250);
    const { keyword, type } = getKeywordAndTypeValues(debouncedInputValue);
    const { data, isLoading, isFetching, isError, error } = useSearch(keyword, type);

    const noOptionsText = getEmptyResultsText(
        isLoading,
        isFetching,
        isError,
        error,
        debouncedInputValue,
        type,
        keyword,
        data
    );

    const searchResultsWithActions = data?.map((result) => {
        const resultInChangelog = changelog.find((member) => member.objectid === result.objectid);
        const matchedSelector = assetGroup.Selectors.find((selector) => selector.selector === result.objectid);

        let action = ChangelogAction.ADD;

        if (result.system_tags?.includes(assetGroup.tag)) {
            action = ChangelogAction.DEFAULT;
        }
        if (matchedSelector) {
            action = ChangelogAction.REMOVE;
        }
        if (resultInChangelog) {
            action = ChangelogAction.UNDO;
        }

        return { ...result, action };
    });

    const handleRenderInput = (params: AutocompleteRenderInputParams): ReactNode => {
        return (
            <TextField
                {...params}
                variant='outlined'
                placeholder={AUTOCOMPLETE_PLACEHOLDER}
                aria-label={AUTOCOMPLETE_PLACEHOLDER}
            />
        );
    };

    const handleRenderOption = (props: HTMLAttributes<HTMLLIElement>, option: AssetGroupChangelogEntry): ReactNode => {
        const actionLabels = {
            [ChangelogAction.ADD]: 'Add',
            [ChangelogAction.REMOVE]: 'Remove',
            [ChangelogAction.DEFAULT]: 'Default Group Member',
            [ChangelogAction.UNDO]: 'Undo',
        };
        return (
            <AutocompleteOption
                key={option.objectid}
                props={props}
                id={option.objectid}
                type={option.type}
                name={option.name}
                actionLabel={actionLabels[option.action]}
            />
        );
    };

    const handleInputChange = (_event: SyntheticEvent, value: string, reason: string): void => {
        if (reason === 'reset') return;
        setSearchInput(value);
    };

    const filterOptions = (options: AssetGroupChangelogEntry[]) => options;
    const getOptionLabel = (option: AssetGroupChangelogEntry) => option.name || option.objectid;
    const getOptionDisabled = (option: AssetGroupChangelogEntry) => option.action === ChangelogAction.DEFAULT;

    return (
        <Autocomplete<any>
            renderInput={handleRenderInput}
            renderOption={handleRenderOption}
            onInputChange={handleInputChange}
            onChange={onChange}
            inputValue={searchInput}
            filterOptions={filterOptions}
            value={null}
            options={searchResultsWithActions || []}
            loading={isLoading || isFetching}
            getOptionLabel={getOptionLabel}
            getOptionDisabled={getOptionDisabled}
            isOptionEqualToValue={() => false}
            clearOnBlur
            clearOnEscape
            disableCloseOnSelect
            noOptionsText={noOptionsText}
            forcePopupIcon={false}
        />
    );
};

export default AssetGroupAutocomplete;
