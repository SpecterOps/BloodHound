import { createFileRoute } from '@tanstack/react-router';
import { authenticateToRoute } from './-utils';

export const Route = createFileRoute('/group-management')({
    beforeLoad: ({ context }) => authenticateToRoute(context.auth),
    staticData: { showNavbar: true },
});
