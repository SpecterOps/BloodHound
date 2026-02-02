import { createLazyFileRoute } from '@tanstack/react-router';
import { Details } from 'bh-shared-ui';

export const Route = createLazyFileRoute('/privilege-zones/zones/$zoneId/rules/$ruleId/objects/$objectId/details')({
    component: Details,
});
