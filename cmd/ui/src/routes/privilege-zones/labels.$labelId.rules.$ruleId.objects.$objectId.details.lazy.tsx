import { createLazyFileRoute } from '@tanstack/react-router';
import { Details } from 'bh-shared-ui';

export const Route = createLazyFileRoute('/privilege-zones/labels/$labelId/rules/$ruleId/objects/$objectId/details')({
    component: Details,
});
