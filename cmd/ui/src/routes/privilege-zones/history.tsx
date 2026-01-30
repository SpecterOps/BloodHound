import { createFileRoute } from '@tanstack/react-router';
import { History } from 'bh-shared-ui';

export const Route = createFileRoute('/privilege-zones/history')({
    component: History,
});
