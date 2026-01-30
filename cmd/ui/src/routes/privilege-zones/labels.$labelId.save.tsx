import { createFileRoute } from '@tanstack/react-router';
import { Save } from 'bh-shared-ui';

export const Route = createFileRoute('/privilege-zones/labels/$labelId/save')({
    component: Save,
});
