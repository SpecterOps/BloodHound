//list of ids to be excluded from Quick Ingest -- useExecuteOnFileDrag
export enum QuickUploadExclusionIds {
    ImportQueryDialog = 'import-query-dialog',
}

export const getExcludedIds = () => {
    const ids = Object.values(QuickUploadExclusionIds);

    for (const id of ids) {
        const element = document.getElementById(id);
        if (element) {
            return true;
        }
    }
    return false;
};
