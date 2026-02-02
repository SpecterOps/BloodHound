import { createLazyFileRoute } from '@tanstack/react-router';
import GroupManagement from 'src/views/GroupManagement/GroupManagement';

export const Route = createLazyFileRoute('/group-management')({
    component: GroupManagement,
});
