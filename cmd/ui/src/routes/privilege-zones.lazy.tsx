import { createLazyFileRoute } from '@tanstack/react-router';
import PrivilegeZones from 'src/views/PrivilegeZones';

export const Route = createLazyFileRoute('/privilege-zones')({
    component: PrivilegeZones,
});
