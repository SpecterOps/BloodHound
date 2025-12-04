import { PageWithTitle } from '../..';
import { SchemaUploadCard } from './SchemaUploadCard';

export const OpenGraphManagement: React.FC = () => {
    return (
        <PageWithTitle
            title='OpenGraph Management'
            pageDescription={
                <p>
                    OpenGraph Management provides a centralized space to define and maintain the structures that shape
                    how BloodHound understands relationships in an environment. Review schema examples in the OpenGraph
                    Library to discover effective modeling patterns.
                </p>
            }>
            <SchemaUploadCard />
        </PageWithTitle>
    );
};
