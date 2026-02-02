import { createFileRoute } from '@tanstack/react-router';
import { authenticateToRoute } from './-utils';

export const Route = createFileRoute('/download-collectors')({
    beforeLoad: ({ context }) => authenticateToRoute(context.auth),
    staticData: { showNavbar: true },
});
