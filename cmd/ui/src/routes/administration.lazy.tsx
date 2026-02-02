import { createLazyFileRoute } from '@tanstack/react-router';
import Administration from 'src/views/Administration';

export const Route = createLazyFileRoute('/administration')({
    component: Administration,
});
