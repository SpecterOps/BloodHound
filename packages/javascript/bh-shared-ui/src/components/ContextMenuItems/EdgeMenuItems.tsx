import { MenuItem } from '@mui/material';
import { FC } from 'react';
import { isEdgeType } from '../../edgeTypes';
import { PathfindingFilters } from '../../hooks';

type Props = {
    id: string;
    pathfindingFilters: PathfindingFilters;
};

const RX_EDGE_TYPE = /^[^_]+_([^_]+)_[^_]+$/;

export const EdgeMenuItems: FC<Props> = ({ id, pathfindingFilters }) => {
    const { handleRemoveEdgeType } = pathfindingFilters;

    const edgeType = id.match(RX_EDGE_TYPE)?.[1];

    const filterEdge = () => {
        if (edgeType) {
            handleRemoveEdgeType(edgeType);
        }
    };

    if (!edgeType) {
        return null;
    }

    // Prevent filtering for edge types not found in AllEdgeTypes array
    return isEdgeType(edgeType) ? (
        <MenuItem key='filter-edge' onClick={filterEdge}>
            Filter out Edge
        </MenuItem>
    ) : (
        <MenuItem key='non-filterable' disabled>
            Non-filterable Edge
        </MenuItem>
    );
};
