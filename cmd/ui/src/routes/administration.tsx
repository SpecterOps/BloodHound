import { createFileRoute, redirect } from '@tanstack/react-router';
import Administration from 'src/views/Administration';
import { ROUTE_LOGIN } from './-constants';

export const Route = createFileRoute('/administration')({
    beforeLoad: ({ context }) => {
        if (context.auth.sessionToken === null && context.auth.user === null) throw redirect({ to: ROUTE_LOGIN });
    },
    component: Administration,
});
