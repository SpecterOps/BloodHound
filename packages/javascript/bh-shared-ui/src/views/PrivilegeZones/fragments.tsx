import { FC } from 'react';

export const RulesLink: FC = () => {
    return (
        <a
            href='https://bloodhound.specterops.io/analyze-data/privilege-zones/selectors'
            target='_blank'
            rel='noopener noreferrer'
            className='text-link underline'>
            Learn more about rules
        </a>
    );
};

export const ZonesLink: FC = () => {
    return (
        <a
            href='https://bloodhound.specterops.io/analyze-data/privilege-zones/zones'
            target='_blank'
            rel='noopener noreferrer'
            className='text-secondary dark:text-secondary-variant-2 underline'>
            Learn more about zones
        </a>
    );
};
