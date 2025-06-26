import {
    Card,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';
import { AppIcon } from '../../components';
import { useGetConfiguration } from '../../hooks';
import { parseTieringConfiguration } from 'js-client-library';

const SalesMessage: FC = () => {

    const { data } = useGetConfiguration();
    const tieringConfig = parseTieringConfiguration(data);

    return tieringConfig && !tieringConfig?.value.multi_tier_analysis_enabled ?
        <Card className='p-3'>
            <CardHeader className='flex flex-row items-center mb-1'>
                <AppIcon.DataAlert size={24} className='mr-2 text-[#ED8537]' />
                <CardTitle>Upgrade Privilege Zones</CardTitle>
            </CardHeader>
            <CardDescription className='p-3 pt-0'>
                <span>
                    You're currently limited to analyzing a single Privilege Zone. Reach out to our{' '}
                    <a
                        href='https://support.bloodhoundenterprise.io/hc/en-us/requests/new'
                        target='_blank'
                        rel='noreferrer'
                        className='text-secondary dark:text-secondary-variant-2 underline'>
                        sales team
                    </a>
                    {' '}to upgrade to analyze and identify risks across multiple zones.
                </span>
            </CardDescription>
        </Card>
        : null
};

export default SalesMessage;