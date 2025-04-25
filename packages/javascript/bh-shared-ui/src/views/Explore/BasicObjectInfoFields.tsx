import { Box } from '@mui/material';
import NodeIcon from '../../components/NodeIcon';
import { ActiveDirectoryNodeKind, AzureNodeKind } from '../../graphSchema';
import { SearchValue } from '../../store';
import { EntityKinds } from '../../utils';
import { Field } from './fragments';

interface BasicObjectInfoFieldsProps {
    handleSourceNodeSelected: (sourceNode: SearchValue) => void;
    objectid: string;
    displayname?: string;
    isTierZero?: boolean;
    isOwned?: boolean;
    service_principal_id?: string;
    noderesourcegroupid?: string;
    grouplinkid?: string;
    name?: string;
}

const RelatedKindField = (
    onSourceNodeSelected: (sourceNode: SearchValue) => void,
    fieldLabel: string,
    relatedKind: EntityKinds,
    id: string,
    name?: string
) => {
    return (
        <Box padding={1}>
            <Box fontWeight='bold' mr={1}>
                {fieldLabel}
            </Box>
            <br />
            <Box display='flex' flexDirection='row' flexWrap='wrap' justifyContent='flex-start'>
                <NodeIcon nodeType={relatedKind} />
                <Box
                    onClick={() => onSourceNodeSelected({ objectid: id, type: relatedKind, name: name || '' })}
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
            {props.isTierZero && <Field label='Tier Zero:' value={true} />}
            {props.isOwned && <Field label='Owned Object:' value={true} />}
            {props.displayname && <Field label='Display Name:' value={props.displayname} />}
            <Field label='Object ID:' value={props.objectid} />
            {props.service_principal_id &&
                RelatedKindField(
                    props.handleSourceNodeSelected,
                    'Service Principal ID:',
                    AzureNodeKind.ServicePrincipal,
                    props.service_principal_id,
                    props.name
                )}
            {props.noderesourcegroupid &&
                RelatedKindField(
                    props.handleSourceNodeSelected,
                    'Node Resource Group ID:',
                    AzureNodeKind.ResourceGroup,
                    props.noderesourcegroupid,
                    props.name
                )}
            {props.grouplinkid &&
                RelatedKindField(
                    props.handleSourceNodeSelected,
                    'Linked Group ID:',
                    ActiveDirectoryNodeKind.Group,
                    props.grouplinkid,
                    props.name
                )}
        </>
    );
};
