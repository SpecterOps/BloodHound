import { createFileRoute } from '@tanstack/react-router';
import { authenticateToRoute } from './-utils';

export const Route = createFileRoute('/administration')({
    beforeLoad: ({ context }) => authenticateToRoute(context.auth),
    staticData: { showNavbar: true },
});
