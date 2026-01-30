import { createFileRoute } from '@tanstack/react-router';
import { ApiExplorer } from 'bh-shared-ui';

export const Route = createFileRoute('/api-explorer')({
    component: ApiExplorer,
});
