import { Autocomplete, AutocompleteRenderInputParams, TextField } from "@mui/material";
import { FC, HTMLAttributes, ReactNode, SyntheticEvent, useState } from "react";
import AutocompleteOption from "./AutocompleteOption";
import { AssetGroupChangelog, AssetGroupChangelogEntry, ChangelogAction } from "./types";
import { AssetGroupMember } from "js-client-library";
import { getKeywordAndTypeValues, useDebouncedValue, useSearch } from "../../hooks";

const AssetGroupAutocomplete: FC<{
    assetGroupMembers: AssetGroupMember[],
    changelog: AssetGroupChangelog,
    onChange: (event: any, value: AssetGroupChangelogEntry) => void,
}> = ({ assetGroupMembers, changelog, onChange }) => {

    const [searchInput, setSearchInput] = useState('');
    const debouncedInputValue = useDebouncedValue(searchInput, 250);
    const { keyword, type } = getKeywordAndTypeValues(debouncedInputValue);
    const { data, isLoading, isFetching } = useSearch(keyword, type);

    const searchResultsWithActions = data?.map(result => {
        const resultInChangelog = changelog.find(member => member.objectid === result.objectid);
        const resultInAssetGroup = assetGroupMembers.find(member => member.object_id === result.objectid);

        let action = ChangelogAction.ADD;
        if (resultInAssetGroup) {
            action = ChangelogAction.DEFAULT;
        }
        if (resultInAssetGroup?.custom_member) {
            action = ChangelogAction.REMOVE;
        }
        if (resultInChangelog) {
            action = ChangelogAction.UNDO;
        }

        return { ...result, action }
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
    }

    const handleRenderOption = (props: HTMLAttributes<HTMLLIElement>, option: AssetGroupChangelogEntry): ReactNode => {
        return (
            <AutocompleteOption
                props={props}
                id={option.objectid}
                type={option.type}
                name={option.name}
                actionLabel={option.action}
            />
        );
    }

    const handleInputChange = (_event: SyntheticEvent, value: string, reason: string): void => {
        if (reason === 'reset') return;
        setSearchInput(value);
    }

    return (
        <Autocomplete<any>
            renderInput={handleRenderInput}
            renderOption={handleRenderOption}
            onInputChange={handleInputChange}
            onChange={onChange}
            inputValue={searchInput}
            filterOptions={(options: AssetGroupChangelogEntry[]) => options}
            value={null}
            options={searchResultsWithActions || []}
            loading={isLoading || isFetching}
            getOptionLabel={(option: AssetGroupChangelogEntry) => option.name || option.objectid}
            isOptionEqualToValue={() => false}
            clearOnBlur
            clearOnEscape
            disableCloseOnSelect
            forcePopupIcon={false}
        />
    )
}

export default AssetGroupAutocomplete;