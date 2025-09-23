import { Button } from '@bloodhoundenterprise/doodleui';
import { useFileUploadDialogContext, usePermissions } from '../../hooks';
import { Permission } from '../../utils';

export const FileIngestUploadButton = () => {
    const { setShowFileIngestDialog } = useFileUploadDialogContext();
    const { checkPermission } = usePermissions();
    const hasPermission = checkPermission(Permission.GRAPH_DB_INGEST);
    const toggleFileUploadDialog = () => setShowFileIngestDialog((prev) => !prev);

    return (
        <Button
            onClick={() => toggleFileUploadDialog()}
            data-testid='file-ingest_button-upload-files'
            disabled={!hasPermission}>
            Upload File(s)
        </Button>
    );
};
