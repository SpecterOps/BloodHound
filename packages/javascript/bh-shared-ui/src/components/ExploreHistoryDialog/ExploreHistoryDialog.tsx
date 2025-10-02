import { Dialog, DialogContent, DialogTitle, DialogTrigger } from '@bloodhoundenterprise/doodleui';
import { CypherEditor } from '@neo4j-cypher/react-codemirror';
import { useSearchParams } from 'react-router-dom';
import { Field, FieldsContainer } from '../../views';
import LabelWithCopy from '../LabelWithCopy';

export const ExploreHistoryDialog = () => {
    const [searchParams] = useSearchParams();

    const htmlFormattedParams = searchParams
        .entries()
        .toArray()
        .map(([key, value]) => {
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
            return (
                <Field key={key} label={<LabelWithCopy label={key} valueToCopy={value} hoverOnly />} value={value} />
            );
        });

    return (
        <Dialog>
            <DialogTrigger>Explore Query Debug</DialogTrigger>
            <DialogContent className='bg-neutral-1'>
                <DialogTitle>Explore Query Debug</DialogTitle>
                <FieldsContainer>{htmlFormattedParams}</FieldsContainer>
            </DialogContent>
        </Dialog>
    );
};
