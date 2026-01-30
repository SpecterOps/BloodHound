import { createFileRoute } from '@tanstack/react-router';
import Administration from 'src/views/Administration';

export const Route = createFileRoute('/administration')({
    component: Administration,
});
