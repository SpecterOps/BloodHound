import { createLazyFileRoute } from '@tanstack/react-router';
import DataQuality from 'src/views/DataQuality';

export const Route = createLazyFileRoute('/administration/data-quality')({
    component: DataQuality,
});
