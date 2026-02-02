import { createLazyFileRoute } from '@tanstack/react-router';
import { Save } from 'bh-shared-ui';

export const Route = createLazyFileRoute('/privilege-zones/labels/$labelId/rules/$ruleId/save')({
    component: Save,
});
