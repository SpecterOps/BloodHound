import { createFileRoute } from '@tanstack/react-router';
import { FileIngest } from 'bh-shared-ui';

export const Route = createFileRoute('/administration/file-ingest')({
    component: FileIngest,
});
