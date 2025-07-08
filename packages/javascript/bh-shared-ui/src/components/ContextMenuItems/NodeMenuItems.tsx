import { MenuItem } from '@mui/material';
import { FC } from 'react';
import { useExploreParams } from '../../hooks';

type Props = {
    objectId: string;
};

export const NodeMenuItems: FC<Props> = ({ objectId }) => {
    const { primarySearch, secondarySearch, setExploreParams } = useExploreParams();

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
