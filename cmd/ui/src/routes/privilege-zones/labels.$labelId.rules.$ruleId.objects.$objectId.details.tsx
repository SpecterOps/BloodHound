import { createFileRoute } from '@tanstack/react-router';
import { Details } from 'bh-shared-ui';

export const Route = createFileRoute('/privilege-zones/labels/$labelId/rules/$ruleId/objects/$objectId/details')({
    component: Details,
});
