import { createFileRoute, redirect } from '@tanstack/react-router';

export const Route = createFileRoute('/administration/')({
    beforeLoad: () => {
        throw redirect({ to: '/administration/file-ingest' });
    },
});
