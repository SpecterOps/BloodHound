import { AssetGroupTag } from 'js-client-library';
import { TagLabelPrefix } from '../useAssetGroupTags';

export const getZoneNameFromKinds = (
    tags: AssetGroupTag[] | undefined,
    kinds: string[] | undefined
): string | undefined => {
    const kindsSet = new Set(kinds);

    const match = tags?.find((tag) => {
        if (tag.type !== 1) return null;
        const underscoredTagName = tag.name.replace(/ /g, '_');
        return kindsSet.has(`${TagLabelPrefix}${underscoredTagName}`);
    });

    return match?.name;
};
