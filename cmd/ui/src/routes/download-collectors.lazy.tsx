import { createLazyFileRoute } from '@tanstack/react-router';
import DownloadCollectors from 'src/views/DownloadCollectors';

export const Route = createLazyFileRoute('/download-collectors')({
    component: DownloadCollectors,
});
