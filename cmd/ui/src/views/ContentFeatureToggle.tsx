import { useFeatureFlag } from 'bh-shared-ui';
import React from 'react';
import { fullyAuthenticatedSelector } from 'src/ducks/auth/authSlice';
import { useAppSelector } from 'src/store';
import Content from './Content';
import ContentV2 from './ContentV2';

const ContentFeatureToggle: React.FC = () => {
    const authState = useAppSelector((state) => state.auth);
    const fullyAuthenticated = useAppSelector(fullyAuthenticatedSelector);

    const { data: flag } = useFeatureFlag('enable_back_button', {
        // block this feature flag check from running on login page
        enabled: !!authState.isInitialized && fullyAuthenticated,
    });

    return flag?.enabled ? <ContentV2 /> : <Content />;
};

export default ContentFeatureToggle;
