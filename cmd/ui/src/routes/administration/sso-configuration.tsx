import { createFileRoute } from '@tanstack/react-router';
import { SSOConfiguration } from 'bh-shared-ui';

export const Route = createFileRoute('/administration/sso-configuration')({
    component: SSOConfiguration,
});
