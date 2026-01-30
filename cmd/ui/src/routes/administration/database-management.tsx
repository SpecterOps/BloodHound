import { createFileRoute } from '@tanstack/react-router';
import DatabaseManagement from 'src/views/DatabaseManagement';

export const Route = createFileRoute('/administration/database-management')({
    component: DatabaseManagement,
});
