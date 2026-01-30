import { createFileRoute } from '@tanstack/react-router';
import DownloadCollectors from 'src/views/DownloadCollectors';

export const Route = createFileRoute('/download-collectors')({
    component: DownloadCollectors,
});
