import { Button } from '@bloodhoundenterprise/doodleui';

export const SchemaUploadCard = () => {
    return (
        <div className='p-6 pr-8 rounded-lg bg-neutral-2 mt-4 max-w-4xl'>
            <h2 className='font-bold text-lg mb-5'>Custom Schema Upload</h2>
            <p className='mb-4'>
                Upload custom schema JSON files to introduce new node and edge types. Then apply and validate schema
                updates to tailor the attack graph model to specific environments, workflows, or needs.
            </p>
            <Button variant='secondary' disabled={true}>
                Upload File
            </Button>
        </div>
    );
};
