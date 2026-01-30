import { createFileRoute } from '@tanstack/react-router';
import PrivilegeZones from 'src/views/PrivilegeZones';

export const Route = createFileRoute('/privilege-zones')({
    component: PrivilegeZones,
});
