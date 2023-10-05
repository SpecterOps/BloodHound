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

export const BasicObjectInfoFields: React.FC<BasicObjectInfoFieldsProps> = (props): JSX.Element => {
    const dispatch = useAppDispatch();
    return (
        <>
            {props.system_tags?.includes(TIER_ZERO_TAG) && <Field label='Tier Zero:' value={true} />}
            {props.displayname && <Field label='Display Name:' value={props.displayname} />}
            <Field label='Object ID:' value={props.objectid} />
            <>
                {props.service_principal_id && (
                    <Box padding={1}>
                        <Box fontWeight='bold' mr={1}>
                            Service Principal ID:
                        </Box>
                        <br />
                        <Box display='flex' flexDirection='row' flexWrap='wrap' justifyContent='flex-start'>
                            <NodeIcon nodeType={GraphNodeTypes.AZServicePrincipal} />
                            <Box
                                data-testid='explore_entity-information-panel-service-principal-id'
                                onClick={() => {
                                    dispatch(
                                        setSearchValue(
                                            {
                                                objectid: props.service_principal_id!,
                                                label: '',
                                                type: GraphNodeTypes.AZServicePrincipal,
                                                name: props.name || '',
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
                                title={props.service_principal_id}>
                                {props.service_principal_id}
                            </Box>
                        </Box>
                    </Box>
                )}
            </>
            <>
                {props.noderesourcegroupid && (
                    <Box padding={1}>
                        <Box fontWeight='bold' mr={1}>
                            Node Resource Group ID:
                        </Box>
                        <br />
                        <Box display='flex' flexDirection='row' flexWrap='wrap' justifyContent='flex-start'>
                            <NodeIcon nodeType={GraphNodeTypes.AZResourceGroup} />
                            <Box
                                data-testid='explore_entity-information-panel-node-resource-group-id'
                                onClick={() => {
                                    dispatch(
                                        setSearchValue(
                                            {
                                                objectid: props.noderesourcegroupid!,
                                                label: '',
                                                type: GraphNodeTypes.AZResourceGroup,
                                                name: props.name || '',
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
                                title={props.noderesourcegroupid}>
                                {props.noderesourcegroupid}
                            </Box>
                        </Box>
                    </Box>
                )}
            </>
        </>
    );
};