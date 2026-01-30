import { createFileRoute } from '@tanstack/react-router';
import { UserProfile } from 'bh-shared-ui';

export const Route = createFileRoute('/my-profile')({
    component: UserProfile,
});
