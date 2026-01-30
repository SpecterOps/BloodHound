import { createFileRoute } from '@tanstack/react-router';
import { Save } from 'bh-shared-ui';

export const Route = createFileRoute('/privilege-zones/zones/$zoneId/save')({
    component: Save,
});
