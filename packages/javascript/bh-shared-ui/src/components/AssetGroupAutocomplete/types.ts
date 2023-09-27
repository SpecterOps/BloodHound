export enum ChangelogAction {
    ADD,
    REMOVE,
}

export type MemberData = {
    objectid: string;
    name: string;
    type: string;
};

export type AssetGroupChangelogEntry = {
    member: MemberData;
    action: ChangelogAction;
};

export type AssetGroupChangelog = AssetGroupChangelogEntry[];
