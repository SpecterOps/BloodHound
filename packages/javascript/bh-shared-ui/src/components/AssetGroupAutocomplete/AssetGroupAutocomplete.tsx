import { Autocomplete, AutocompleteRenderInputParams, TextField } from "@mui/material";
import { FC, HTMLAttributes, ReactNode, SyntheticEvent } from "react";
import AutocompleteOption from "./AutocompleteOption";
import { AssetGroupChangelog, MemberData } from ".";
import { UseQueryResult } from "react-query";

const AssetGroupAutocomplete: FC<{
    search: UseQueryResult<MemberData[], any>,
    changelog: AssetGroupChangelog,
    inputValue: string,
    onInputChange: (event: any, value: string) => void,
}> = ({ search, inputValue, changelog, onInputChange }) => {

    const { data, error, isError, isLoading, isFetching } = search;

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

    const handleRenderOption = (props: HTMLAttributes<HTMLLIElement>, option: MemberData): ReactNode => {

        // const memberInChangelog = changelog.filter(change => change.member.object_id === option.object_id)[0];

        // Default is to display "Add" if node is not in asset group
        // If it is already in the asset group and is a custom member, display "Remove"
        // If the member is not custom member, we need to display "Default Group Member"
        // If the member is already in the changelog, we display "Undo" followed by the action value, which should be a string

        return (
            <AutocompleteOption
                props={props}
                id={option.objectid}
                type={option.type}
                name={option.name}
                actionLabel="add"
            />
        );
    }

    const handleInputChange = (_event: SyntheticEvent, value: string, reason: string): void => {
        if (reason !== 'reset') onInputChange(_event, value);
    }

    return (
        <Autocomplete<any>
            renderInput={handleRenderInput}
            renderOption={handleRenderOption}
            onInputChange={handleInputChange}
            inputValue={inputValue}
            filterOptions={(options: MemberData[]) => options}
            value={null}
            options={data || []}
            loading={isLoading || isFetching}
            getOptionLabel={(option: MemberData) => option.name || option.objectid}
            isOptionEqualToValue={() => false}
            clearOnBlur
            clearOnEscape
            disableCloseOnSelect
            forcePopupIcon={false}
        />
    )
}

export default AssetGroupAutocomplete;