//list of ids to be excluded from Quick Ingest -- useExecuteOnFileDrag
const quickUploadExclusions = ['import-query-dialog'];

export const getExcludedIds = () => {
    for (const id of quickUploadExclusions) {
        const element = document.getElementById(id);
        if (element) {
            return true;
        }
    }
    return false;
};
