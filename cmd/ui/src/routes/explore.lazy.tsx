import { createLazyFileRoute } from '@tanstack/react-router';
import GraphView from 'src/views/Explore/GraphView';

export const Route = createLazyFileRoute('/explore')({ component: GraphView });
