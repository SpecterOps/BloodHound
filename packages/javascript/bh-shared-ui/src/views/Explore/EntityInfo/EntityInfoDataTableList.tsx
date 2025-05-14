import { Box, Divider } from '@mui/material';
import { ActiveDirectoryNodeKind } from '../../../graphSchema';
import { allSections, EntityKinds } from '../../../utils';
import React from 'react';
import { EntityInfoContentProps } from './EntityInfoContent';
import EntityInfoDataTable from './EntityInfoDataTable';

const EntityInfoDataTableList: React.FC<EntityInfoContentProps> = ({ id, nodeType }) => {
    let type = nodeType as EntityKinds;
    if (nodeType === ActiveDirectoryNodeKind.LocalGroup || nodeType === ActiveDirectoryNodeKind.LocalUser)
        type = ActiveDirectoryNodeKind.Entity;
    const tables = allSections[type]?.(id) || [];

    return (
        <>
            {tables.map((table, index) => (
                <React.Fragment key={index}>
                    <Box padding={1}>
                        <Divider />
                    </Box>
                    <EntityInfoDataTable {...table} />
                </React.Fragment>
            ))}
        </>
    );
};

export default EntityInfoDataTableList;
