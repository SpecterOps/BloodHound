import { Autocomplete, AutocompleteRenderInputParams, TextField } from '@mui/material';
import { FC, HTMLAttributes, ReactNode, SyntheticEvent, useState } from 'react';
import AutocompleteOption from './AutocompleteOption';
import { AssetGroupChangelog, AssetGroupChangelogEntry, ChangelogAction } from './types';
import { AssetGroup } from 'js-client-library';
import { getEmptyResultsText, getKeywordAndTypeValues, useDebouncedValue, useSearch } from '../../hooks';

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

        console.log(assetGroup.Selectors, result.objectid);

        let action = ChangelogAction.ADD;

        if (result.system_tags.includes(assetGroup.tag)) {
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
                placeholder='Add or Remove Members'
                aria-label='Add or Remove Members'
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
