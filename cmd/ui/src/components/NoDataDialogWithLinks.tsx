import { Link } from 'react-router-dom';
import { NoDataDialog } from 'bh-shared-ui';
import { ROUTE_ADMINISTRATION_FILE_INGEST } from 'src/ducks/global/routes';

type NoDataDialogWithLinksProps = {
    open: boolean;
};

const linkStyles = 'text-secondary dark:text-secondary-variant-2 hover:underline';

const fileIngestLinkProps = {
    className: linkStyles,
    to: ROUTE_ADMINISTRATION_FILE_INGEST,
};

const gettingStartedLinkProps = {
    className: linkStyles,
    target: '_blank',
    rel: 'noreferrer',
    href: 'https://support.bloodhoundenterprise.io/hc/en-us/sections/17274904083483-BloodHound-CE-Collection',
};

const sampleCollectionLinkProps = {
    className: linkStyles,
    target: '_blank',
    rel: 'noreferrer',
    href: 'https://github.com/SpecterOps/BloodHound/wiki/Example-Data',
};

export const NoDataDialogWithLinks: React.FC<NoDataDialogWithLinksProps> = ({ open }) => {
    return (
        <NoDataDialog open={open}>
            To explore your environment, <Link {...fileIngestLinkProps}>start by uploading your data</Link> on the file
            ingest page.
            <br className='mb-4' />
            Need help? Check out the <a {...gettingStartedLinkProps}>Getting Started guide</a> for instructions.
            <br className='mb-4' />
            If you want to test BloodHound with sample data, you may download some from our{' '}
            <a {...sampleCollectionLinkProps}>Sample Collection</a> GitHub page.
        </NoDataDialog>
    );
};
