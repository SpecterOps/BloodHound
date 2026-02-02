import { createFileRoute, redirect } from '@tanstack/react-router';
import { ROUTE_EXPLORE } from './-constants';

// redirect any unmatched wildcard path to the home/explore page
export const Route = createFileRoute('/$')({
    beforeLoad: () => {
        throw redirect({ to: ROUTE_EXPLORE, replace: true });
    },
});
