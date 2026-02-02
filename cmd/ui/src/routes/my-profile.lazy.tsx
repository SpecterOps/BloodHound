import { createLazyFileRoute } from '@tanstack/react-router';
import { UserProfile } from 'bh-shared-ui';

export const Route = createLazyFileRoute('/my-profile')({
    component: UserProfile,
});
