import { createLazyFileRoute } from '@tanstack/react-router';
import { Save } from 'bh-shared-ui';

export const Route = createLazyFileRoute('/privilege-zones/zones/$zoneId/rules/save')({
    component: Save,
});
