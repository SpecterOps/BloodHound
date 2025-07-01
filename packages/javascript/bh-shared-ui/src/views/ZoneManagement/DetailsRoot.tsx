import { FC } from 'react';
import { AppNavigate } from '../../components';
import { useHighestPrivilegeTag } from '../../hooks';
import { DEFAULT_ZONE_MANAGEMENT_ROUTE } from '../../routes';

const DetailsRoot: FC = () => {
    const topTagId = useHighestPrivilegeTag()?.id;
    return <AppNavigate to={'/zone-management/' + DEFAULT_ZONE_MANAGEMENT_ROUTE + topTagId} replace />;
};

export default DetailsRoot;
