import { useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import { setParamsFactory } from '../../utils';
import { useHighestPrivilegeTag } from '../useAssetGroupTags';

export type ZoneQueryParams = {
    assetGroupTagId: number | undefined;
};

const parseAssetGroupTagId = (assetGroupTagId: string | null, topTagId: number | undefined): number | undefined => {
    if (assetGroupTagId !== null) {
        return parseInt(assetGroupTagId);
    }

    return topTagId;
};

export const useZoneQueryParams = () => {
    const [searchParams, setSearchParams] = useSearchParams();
    const topTagId = useHighestPrivilegeTag()?.id;

    const assetGroupTagId = parseAssetGroupTagId(searchParams.get('assetGroupTagId'), topTagId);

    const params = new URLSearchParams();
    if (typeof assetGroupTagId === 'number') params.append('asset_group_tag_id', assetGroupTagId.toString());

    return {
        assetGroupTagId,
        params,
        setZoneQueryParams: useCallback(
            (updatedParams: Partial<ZoneQueryParams>) =>
                setParamsFactory(setSearchParams, ['assetGroupTagId'])(updatedParams),
            [setSearchParams]
        ),
    };
};
