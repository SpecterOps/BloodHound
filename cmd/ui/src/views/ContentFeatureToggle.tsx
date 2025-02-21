import { useFeatureFlag } from 'bh-shared-ui';
import React from 'react';
import Content from './Content';
import ContentV2 from './ContentV2';

const ContentFeatureToggle: React.FC = () => {
    const { data: flag } = useFeatureFlag('back_button_support');

    return flag?.enabled ? <ContentV2 /> : <Content />;
};

export default ContentFeatureToggle;
