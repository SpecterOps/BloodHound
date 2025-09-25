import { SeedTypeObjectId } from 'js-client-library';
import { useMutation } from 'react-query';
import { useNotifications } from '../providers';
import { detailsPath, labelsPath, privilegeZonesPath, zonesPath } from '../routes';
import { Permission, apiClient } from '../utils';
import { AssetGroupMenuItem } from '../views';
import { useTagsQuery } from './useAssetGroupTags';
import { useExploreSelectedItem } from './useExploreSelectedItem';
import { NodeResponse, isNode } from './useGraphItem';
import { usePermissions } from './usePermissions';

const useAssetGroupMenuItems = (setShowContextMenu: React.Dispatch<React.SetStateAction<boolean>>) => {
    const { checkPermission } = usePermissions();
    const { addNotification } = useNotifications();
    const getAssetGroupTagsQuery = useTagsQuery();
    const { selectedItemQuery } = useExploreSelectedItem();

    const createAssetGroupTagSelectorMutation = useMutation({
        mutationFn: ({ assetGroupId, node }: { assetGroupId: string | number; node: NodeResponse }) => {
            return apiClient.createAssetGroupTagSelector(assetGroupId, {
                name: node.label ?? node.objectId,
                seeds: [
                    {
                        type: SeedTypeObjectId,
                        value: node.objectId,
                    },
                ],
            });
        },
        onSuccess: () => {
            addNotification('Node successfully added.', 'AssetGroupUpdateSuccess');
        },
        onError: (error: any) => {
            console.error(error);
            addNotification('An error occurred when adding node', 'AssetGroupUpdateError');
        },
    });

    const handleAddNode = (assetGroupId: string | number) => {
        if (!createAssetGroupTagSelectorMutation.isLoading) {
            createAssetGroupTagSelectorMutation.mutate({
                assetGroupId,
                node: selectedItemQuery.data as NodeResponse,
            });
        }
    };

    const tierZeroAssetGroup = getAssetGroupTagsQuery.data?.find((value) => {
        return value.position === 1;
    });

    const ownedAssetGroup = getAssetGroupTagsQuery.data?.find((value) => {
        return value.type === 3;
    });

    if (!checkPermission(Permission.GRAPH_DB_WRITE)) return [];

    return [
        ...(tierZeroAssetGroup
            ? [
                  <AssetGroupMenuItem
                      key={tierZeroAssetGroup.id}
                      disableAddNode={createAssetGroupTagSelectorMutation.isLoading}
                      assetGroupId={tierZeroAssetGroup.id}
                      assetGroupName={tierZeroAssetGroup.name}
                      onAddNode={handleAddNode}
                      removeNodePath={`/${privilegeZonesPath}/${zonesPath}/${tierZeroAssetGroup.id}/${detailsPath}`}
                      isCurrentMember={isNode(selectedItemQuery.data) && selectedItemQuery.data.isTierZero}
                      onShowConfirmation={() => {
                          setShowContextMenu(true);
                      }}
                      onCancelConfirmation={() => {
                          setShowContextMenu(false);
                      }}
                      showConfirmationOnAdd
                      confirmationOnAddMessage={`Are you sure you want to add this node to ${tierZeroAssetGroup.name}? This action will initiate an analysis run to update group membership.`}
                  />,
              ]
            : []),
        ...(ownedAssetGroup
            ? [
                  <AssetGroupMenuItem
                      key={ownedAssetGroup.id}
                      disableAddNode={createAssetGroupTagSelectorMutation.isLoading}
                      assetGroupId={ownedAssetGroup.id}
                      assetGroupName={ownedAssetGroup.name}
                      onAddNode={handleAddNode}
                      removeNodePath={`/${privilegeZonesPath}/${labelsPath}/${ownedAssetGroup.id}/${detailsPath}`}
                      isCurrentMember={isNode(selectedItemQuery.data) && selectedItemQuery.data.isOwnedObject}
                  />,
              ]
            : []),
    ];
};

export default useAssetGroupMenuItems;
