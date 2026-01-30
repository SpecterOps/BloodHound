import { createFileRoute } from '@tanstack/react-router';
import BloodHoundConfiguration from 'src/views/BloodHoundConfiguration';

export const Route = createFileRoute('/administration/bloodhound-configuration')({
    component: BloodHoundConfiguration,
});
