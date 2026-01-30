import { createFileRoute, redirect } from '@tanstack/react-router';

// redirect any unmatched wildcard path to the home/explore page
export const Route = createFileRoute('/$')({
    beforeLoad: () => {
        throw redirect({ to: '/explore', replace: true });
    },
});
