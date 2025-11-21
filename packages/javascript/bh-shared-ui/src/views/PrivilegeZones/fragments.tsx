import { FC, useContext } from 'react';
import { PrivilegeZonesContext } from './PrivilegeZonesContext';

export const RulesLink: FC = () => {
    return (
        <a
            href='https://bloodhound.specterops.io/analyze-data/privilege-zones/selectors'
            target='_blank'
            rel='noopener noreferrer'
            className='text-link underline'>
            Learn more about Rules
        </a>
    );
};

export const ZonesLink: FC = () => {
    return (
        <a
            href='https://bloodhound.specterops.io/analyze-data/privilege-zones/zones'
            target='_blank'
            rel='noopener noreferrer'
            className='text-link underline'>
            Learn more about Zones
        </a>
    );
};

export const PageDescription: FC = () => {
    const { SupportLink } = useContext(PrivilegeZonesContext);

    return (
        <p className='mt-6'>
            Use Privilege Zones to segment and organize assets based on sensitivity and access level.
            <br />
            Learn about{' '}
            <a
                href='https://bloodhound.specterops.io/analyze-data/privilege-zones/overview'
                target='_blank'
                rel='noopener noreferrer'
                className='text-link underline'>
                setup and best practices
            </a>
            . <span>{SupportLink && <SupportLink />}</span>
        </p>
    );
};
