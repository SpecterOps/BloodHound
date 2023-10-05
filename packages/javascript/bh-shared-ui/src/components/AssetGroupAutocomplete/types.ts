export enum ChangelogAction {
    ADD = 'Add',
    REMOVE = 'Remove',
    DEFAULT = 'Default Group Member',
    UNDO = 'Undo',
}

export type MemberData = {
    objectid: string;
    name: string;
    type: string;
};

export type AssetGroupChangelogEntry = MemberData & { action: ChangelogAction };

export type AssetGroupChangelog = AssetGroupChangelogEntry[];
