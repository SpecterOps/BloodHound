import { createLazyFileRoute } from '@tanstack/react-router';
import EarlyAccessFeatures from 'src/views/EarlyAccessFeatures';

export const Route = createLazyFileRoute('/administration/early-access-features')({
    component: EarlyAccessFeatures,
});
