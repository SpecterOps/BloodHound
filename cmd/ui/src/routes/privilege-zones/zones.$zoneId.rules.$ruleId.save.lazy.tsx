import { createLazyFileRoute } from '@tanstack/react-router';
import { Save } from 'bh-shared-ui';

export const Route = createLazyFileRoute('/privilege-zones/zones/$zoneId/rules/$ruleId/save')({
    component: Save,
});
