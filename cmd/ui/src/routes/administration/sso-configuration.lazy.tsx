import { createLazyFileRoute } from '@tanstack/react-router';
import { SSOConfiguration } from 'bh-shared-ui';

export const Route = createLazyFileRoute('/administration/sso-configuration')({
    component: SSOConfiguration,
});
