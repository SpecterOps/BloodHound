import { MenuItem } from '@mui/material';
import { FC } from 'react';
import { useExploreParams } from '../../hooks';

type Props = {
    exploreParams: ReturnType<typeof useExploreParams>;
    objectId: string;
};

export const NodeMenuItems: FC<Props> = ({ exploreParams, objectId }) => {
    const { primarySearch, secondarySearch, setExploreParams } = exploreParams;

    return (
        <>
            <MenuItem
                key='starting-node'
                onClick={() =>
                    setExploreParams({
                        exploreSearchTab: 'pathfinding',
                        searchType: secondarySearch ? 'pathfinding' : 'node',
                        primarySearch: objectId,
                    })
                }>
                Set as starting node
            </MenuItem>

            <MenuItem
                key='ending-node'
                onClick={() =>
                    setExploreParams({
                        exploreSearchTab: 'pathfinding',
                        searchType: primarySearch ? 'pathfinding' : 'node',
                        secondarySearch: objectId,
                    })
                }>
                Set as ending node
            </MenuItem>
        </>
    );
};
