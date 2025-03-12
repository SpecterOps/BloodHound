import { searchbarActions, useExploreParams, useFeatureFlag } from 'bh-shared-ui';
import { useAppDispatch, useAppSelector } from 'src/store';

const useContextMenuStateSwitch = (contextMenuNodeId?: string) => {
    const { data: flag } = useFeatureFlag('back_button_support');

    const dispatch = useAppDispatch();
    const selectedNodeFromRedux = useAppSelector((state) => state.entityinfo.selectedNode);

    // context menu id could be derived not from redux
    const { primarySearch, secondarySearch, setExploreParams } = useExploreParams();

    if (flag?.enabled) {
        return {
            handleSetStartingNode: () => {
                if (contextMenuNodeId) {
                    const searchType = secondarySearch ? 'pathfinding' : 'node';
                    setExploreParams({
                        exploreSearchTab: 'pathfinding',
                        searchType,
                        primarySearch: contextMenuNodeId,
                    });
                }
            },
            handleSetEndingNode: () => {
                const searchType = primarySearch ? 'pathfinding' : 'node';
                if (contextMenuNodeId) {
                    setExploreParams({
                        exploreSearchTab: 'pathfinding',
                        searchType,
                        secondarySearch: contextMenuNodeId,
                    });
                }
            },
        };
    } else {
        return {
            handleSetStartingNode: () => {
                if (selectedNodeFromRedux) {
                    dispatch(searchbarActions.tabChanged('secondary'));
                    dispatch(
                        searchbarActions.sourceNodeSelected(
                            {
                                name: selectedNodeFromRedux.name,
                                objectid: selectedNodeFromRedux.id,
                                type: selectedNodeFromRedux.type,
                            },
                            true
                        )
                    );
                }
            },
            handleSetEndingNode: () => {
                if (selectedNodeFromRedux) {
                    dispatch(searchbarActions.tabChanged('secondary'));
                    dispatch(
                        searchbarActions.destinationNodeSelected({
                            name: selectedNodeFromRedux.name,
                            objectid: selectedNodeFromRedux.id,
                            type: selectedNodeFromRedux.type,
                        })
                    );
                }
            },
        };
    }
};

export default useContextMenuStateSwitch;
