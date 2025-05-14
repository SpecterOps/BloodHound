import { Box } from '@mui/material';
import { EntityKinds } from '../../../utils';
import React from 'react';
import EntityInfoDataTableList from './EntityInfoDataTableList';
import EntityObjectInformation from './EntityObjectInformation';

export interface EntityInfoContentProps {
    id: string;
    nodeType: EntityKinds | string;
    databaseId?: string;
}

const EntityInfoContent: React.FC<EntityInfoContentProps> = (props) => {
    return (
        <Box>
            <EntityObjectInformation {...props} />
            <EntityInfoDataTableList {...props} />
        </Box>
    );
};

export default EntityInfoContent;
