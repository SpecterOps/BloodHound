import { Dialog, DialogContent, DialogTitle, DialogTrigger } from '@bloodhoundenterprise/doodleui';
import { useSearchParams } from 'react-router-dom';

export const ExploreHistoryDialog = () => {
    const [searchParams] = useSearchParams();

    const formattedParams = searchParams.entries().map(([key, value]) => {
        if (key === 'cypherSearch') {
            return [key, atob(value)];
        } else {
            return [key, value];
        }
    });
    return (
        <Dialog>
            <DialogTrigger>Explore Query Debug</DialogTrigger>
            <DialogContent className='bg-neutral-1'>
                <DialogTitle>Explore Query Debug</DialogTitle>
                <pre style={{ overflowX: 'auto', whiteSpace: 'pre-wrap', wordWrap: 'break-word' }}>
                    {JSON.stringify(Object.fromEntries(formattedParams), null, 2)}
                </pre>
            </DialogContent>
        </Dialog>
    );
};
