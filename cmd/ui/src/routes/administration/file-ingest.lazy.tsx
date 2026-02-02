import { createLazyFileRoute } from '@tanstack/react-router';
import { FileIngest } from 'bh-shared-ui';

export const Route = createLazyFileRoute('/administration/file-ingest')({
    component: FileIngest,
});
