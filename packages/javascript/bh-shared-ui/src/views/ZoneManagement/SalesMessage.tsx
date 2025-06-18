import {
    Card,
    CardContent,
    CardHeader,
    CardTitle,
} from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';
import { AppIcon } from '../../components';

const SalesMessage: FC = () => {

    return (
        <Card className='p-3'>
            <CardHeader className='flex flex-row items-end'>
                <AppIcon.DataAlert size={24} className='mr-2 text-[#ED8537]' />
                <CardTitle>Upgrade Priviledge Zones</CardTitle>
            </CardHeader>
            <CardContent>
                <p>
                    You're currently limited to analyzing a single Privilege Zone. Reach out to our sales team to upgrade to analyze and identify risks across multiple zones.
                </p>
            </CardContent>
        </Card>
    )
};

export default SalesMessage;