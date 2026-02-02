import { createLazyFileRoute } from '@tanstack/react-router';
import { History } from 'bh-shared-ui';

export const Route = createLazyFileRoute('/privilege-zones/history')({
    component: History,
});
