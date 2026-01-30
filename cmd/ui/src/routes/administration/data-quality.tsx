import { createFileRoute } from '@tanstack/react-router';
import DataQuality from 'src/views/DataQuality';

export const Route = createFileRoute('/administration/data-quality')({
    component: DataQuality,
});
