import { createFileRoute } from '@tanstack/react-router';
import EarlyAccessFeatures from 'src/views/EarlyAccessFeatures';

export const Route = createFileRoute('/administration/early-access-features')({
    component: EarlyAccessFeatures,
});
