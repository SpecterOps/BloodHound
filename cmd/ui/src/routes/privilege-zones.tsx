import { createFileRoute } from '@tanstack/react-router';
import { authenticateToRoute } from './-utils';

export const Route = createFileRoute('/privilege-zones')({
    beforeLoad: ({ context }) => authenticateToRoute(context.auth),
});
