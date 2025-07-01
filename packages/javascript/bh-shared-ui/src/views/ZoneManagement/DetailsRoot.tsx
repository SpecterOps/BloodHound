import { Skeleton } from '@mui/material';
import { FC } from 'react';
import { AppNavigate } from '../../components';
import { useHighestPrivilegeTag } from '../../hooks';
import { DEFAULT_ZONE_MANAGEMENT_ROUTE } from '../../routes';
const DetailsRoot: FC = () => {
    const topTagId = useHighestPrivilegeTag()?.id;
    if (topTagId) {
        return <AppNavigate to={'/zone-management/' + DEFAULT_ZONE_MANAGEMENT_ROUTE + topTagId} replace />;
    } else {
        return <Skeleton className='h-24' />;
    }
};

export default DetailsRoot;
