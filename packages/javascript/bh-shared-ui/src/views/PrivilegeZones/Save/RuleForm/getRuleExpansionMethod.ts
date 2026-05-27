import {
    SeedExpansionMethod,
    SeedExpansionMethodAll,
    SeedExpansionMethodChild,
    SeedExpansionMethodNone,
} from 'js-client-library';

export const getRuleExpansionMethod = (
    tagId: string,
    tagType: 'labels' | 'zones',
    ownedId: string | undefined
): SeedExpansionMethod => {
    // Owned is a specific tag type that does not undergo expansion
    if (tagId === ownedId) return SeedExpansionMethodNone;

    return tagType === 'zones' ? SeedExpansionMethodAll : SeedExpansionMethodChild;
};
