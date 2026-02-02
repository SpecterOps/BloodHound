import { createLazyFileRoute } from '@tanstack/react-router';
import BloodHoundConfiguration from 'src/views/BloodHoundConfiguration';

export const Route = createLazyFileRoute('/administration/bloodhound-configuration')({
    component: BloodHoundConfiguration,
});
