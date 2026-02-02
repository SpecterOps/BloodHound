import { createFileRoute, redirect } from '@tanstack/react-router';

export const Route = createFileRoute('/privilege-zones/')({
    beforeLoad: () => {
        throw redirect({ to: '/privilege-zones/zones/$zoneId/details', params: { zoneId: '1' } });
    },
});
