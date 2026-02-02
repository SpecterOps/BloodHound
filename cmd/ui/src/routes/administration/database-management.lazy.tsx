import { createLazyFileRoute } from '@tanstack/react-router';
import DatabaseManagement from 'src/views/DatabaseManagement';

export const Route = createLazyFileRoute('/administration/database-management')({
    component: DatabaseManagement,
});
