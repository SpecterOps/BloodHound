import { createLazyFileRoute } from '@tanstack/react-router';
import { Users } from 'bh-shared-ui';

export const Route = createLazyFileRoute('/administration/manage-users')({
    component: Users,
});
