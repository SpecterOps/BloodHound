import { createFileRoute } from '@tanstack/react-router';
import OpenGraphManagement from 'bh-shared-ui/OpenGraphManagement';

export const Route = createFileRoute('/administration/opengraph-management')({
    component: OpenGraphManagement,
});
