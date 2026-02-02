import { createLazyFileRoute } from '@tanstack/react-router';
import { ApiExplorer } from 'bh-shared-ui';

export const Route = createLazyFileRoute('/api-explorer')({
    component: ApiExplorer,
});
