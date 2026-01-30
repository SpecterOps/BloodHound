import { createFileRoute } from '@tanstack/react-router';
import GraphView from 'src/views/Explore/GraphView';

export const Route = createFileRoute('/explore')({
    component: GraphView,
});
