import React from 'react';
import { useFeatureFlag } from '../../hooks/useFeatureFlags';
import GroupManagementContent from './GroupManagementContent';
import GroupManagementContentV2, { GroupManagementContentV2Props } from './GroupManagementContentV2';

const GroupManagementFeatureToggle: React.FC<GroupManagementContentV2Props> = (props) => {
    const { data: flag } = useFeatureFlag('back_button_support');
    return flag?.enabled ? <GroupManagementContentV2 {...props} /> : <GroupManagementContent {...props} />;
};

export default GroupManagementFeatureToggle;
