export enum ChangelogAction {
    ADD = 'add',
    REMOVE = 'remove',
}

export type MemberData = {
    objectid: string;
    name: string;
    type: string;
};

export type AssetGroupChangelogEntry = MemberData & { action: ChangelogAction };

export type AssetGroupChangelog = AssetGroupChangelogEntry[];
