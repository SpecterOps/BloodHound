import { createFileRoute } from '@tanstack/react-router';
import { Users } from 'bh-shared-ui';

export const Route = createFileRoute('/administration/manage-users')({
    component: Users,
});
