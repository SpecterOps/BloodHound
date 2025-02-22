import { useFeatureFlag } from 'bh-shared-ui';
import React from 'react';
import GroupManagement from './GroupManagement';
import GroupManagementV2 from './GroupManagementV2';

const GroupManagementFeatureToggle: React.FC = () => {
    const { data: flag } = useFeatureFlag('back_button_support');

    return flag?.enabled ? <GroupManagementV2 /> : <GroupManagement />;
};

export default GroupManagementFeatureToggle;
