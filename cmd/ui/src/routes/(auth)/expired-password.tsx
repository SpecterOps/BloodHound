import { createFileRoute } from '@tanstack/react-router';
import { authenticateToRoute } from '../-utils';

export const Route = createFileRoute('/(auth)/expired-password')({
    beforeLoad: ({ context }) => authenticateToRoute(context.auth),
    staticData: { showNavbar: false },
});
