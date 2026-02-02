import { createLazyFileRoute } from '@tanstack/react-router';
import { Details } from 'bh-shared-ui';

export const Route = createLazyFileRoute('/privilege-zones/zones/$zoneId/details')({
    component: Details,
});
