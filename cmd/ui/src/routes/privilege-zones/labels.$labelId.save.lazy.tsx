import { createLazyFileRoute } from '@tanstack/react-router';
import { Save } from 'bh-shared-ui';

export const Route = createLazyFileRoute('/privilege-zones/labels/$labelId/save')({
    component: Save,
});
