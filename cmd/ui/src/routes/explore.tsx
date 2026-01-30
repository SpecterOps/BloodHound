import { createFileRoute } from '@tanstack/react-router';
import { authenticateToRoute } from './-utils';

export const Route = createFileRoute('/explore')({ beforeLoad: ({ context }) => authenticateToRoute(context.auth) });
