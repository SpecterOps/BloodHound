import { Box } from "@mui/material";
import { NodeIcon, Field } from "bh-shared-ui";
import { TIER_ZERO_TAG } from "src/constants";
import { GraphNodeTypes } from "src/ducks/graph/types";
import { setSearchValue, startSearchSelected } from "src/ducks/searchbar/actions";
import { PRIMARY_SEARCH, SEARCH_TYPE_EXACT } from "src/ducks/searchbar/types";
import { useAppDispatch } from "src/store";

interface BasicObjectInfoFieldsProps {
    objectid: string;
    displayname?: string;
    system_tags?: string;
    service_principal_id?: string;
    noderesourcegroupid?: string;
    name?: string;
}

const RelatedKindField = (fieldLabel: string, relatedKind: GraphNodeTypes, id: string, name?: string) => {
    const dispatch = useAppDispatch();
    return (
        <Box padding={1}>
            <Box fontWeight='bold' mr={1}>
                {fieldLabel}
            </Box>
            <br />
            <Box display='flex' flexDirection='row' flexWrap='wrap' justifyContent='flex-start'>
                <NodeIcon nodeType={relatedKind} />
                <Box
                    onClick={() => {
                        dispatch(
                            setSearchValue(
                                {
                                    objectid: id,
                                    label: '',
                                    type: relatedKind,
                                    name: name || '',
                                },
                                PRIMARY_SEARCH,
                                SEARCH_TYPE_EXACT
                            )
                        );
                        dispatch(startSearchSelected(PRIMARY_SEARCH));
                    }}
                    style={{ cursor: 'pointer' }}
                    overflow='hidden'
                    textOverflow='ellipsis'
                    title={id}>
                    {id}
                </Box>
            </Box>
        </Box>
    );
};

export const BasicObjectInfoFields: React.FC<BasicObjectInfoFieldsProps> = (props): JSX.Element => {
    return (
        <>
            {props.system_tags?.includes(TIER_ZERO_TAG) && <Field label='Tier Zero:' value={true} />}
            {props.displayname && <Field label='Display Name:' value={props.displayname} />}
            <Field label='Object ID:' value={props.objectid} />
            {props.service_principal_id &&
                RelatedKindField(
                    'Service Principal ID:',
                    GraphNodeTypes.AZServicePrincipal,
                    props.service_principal_id,
                    props.name
                )}
            {props.noderesourcegroupid &&
                RelatedKindField(
                    'Node Resource Group ID:',
                    GraphNodeTypes.AZResourceGroup,
                    props.noderesourcegroupid,
                    props.name
                )}
        </>
    );
}