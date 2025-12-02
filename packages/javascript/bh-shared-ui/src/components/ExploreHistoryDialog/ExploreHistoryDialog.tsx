import { Dialog, DialogContent, DialogDescription, DialogTitle, DialogTrigger } from '@bloodhoundenterprise/doodleui';
import { CypherEditor } from '@neo4j-cypher/react-codemirror';
import { useSearchParams } from 'react-router-dom';
import { Field, FieldsContainer } from '../../views/Explore/fragments';
import LabelWithCopy from '../LabelWithCopy';

export const ExploreHistoryDialog = () => {
    const [searchParams] = useSearchParams();

    const paramArray = Array.from(searchParams.entries());

    const formattedParams = paramArray.map(([key, value]) => {
        if (key === 'cypherSearch') {
            return (
                <div key={key} className={'my-2'}>
                    <strong>
                        <LabelWithCopy label={key} valueToCopy={value} hoverOnly />
                    </strong>
                    <CypherEditor
                        value={atob(value)}
                        readOnly={true}
                        theme={document.documentElement.classList.contains('dark') ? 'dark' : 'light'}
                    />
                </div>
            );
        }
        return <Field key={key} label={<LabelWithCopy label={key} valueToCopy={value} hoverOnly />} value={value} />;
    });

    return (
        <Dialog>
            <DialogTrigger className='w-full text-left'>Explore Query Debug</DialogTrigger>
            <DialogContent className='bg-neutral-1'>
                <DialogTitle>Explore Query Debug</DialogTitle>
                <DialogDescription>
                    Collection of all query params representing the current state of the Explore page.
                </DialogDescription>
                <FieldsContainer>{formattedParams}</FieldsContainer>
            </DialogContent>
        </Dialog>
    );
};
