import { createFileRoute, redirect } from '@tanstack/react-router';

export const Route = createFileRoute('/privilege-zones/')({
    beforeLoad: () => {
        throw redirect({ to: '/privilege-zones/zones/1/details' });
    },
});
