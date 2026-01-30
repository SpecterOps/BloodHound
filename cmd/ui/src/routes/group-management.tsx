import { createFileRoute } from '@tanstack/react-router';
import GroupManagement from 'src/views/GroupManagement/GroupManagement';

export const Route = createFileRoute('/group-management')({
    component: GroupManagement,
});
