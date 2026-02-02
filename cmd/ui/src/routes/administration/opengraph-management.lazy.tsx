import { createLazyFileRoute } from '@tanstack/react-router';
import OpenGraphManagement from 'bh-shared-ui/OpenGraphManagement';

export const Route = createLazyFileRoute('/administration/opengraph-management')({
    component: OpenGraphManagement,
});
