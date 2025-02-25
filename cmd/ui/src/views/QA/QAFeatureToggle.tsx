import { useFeatureFlag } from 'bh-shared-ui';
import React from 'react';
import QualityAssurance from './QA';
import QualityAssuranceV2 from './QAV2';

const QAFeatureToggle: React.FC = () => {
    const { data: flag } = useFeatureFlag('back_button_support');
    return flag?.enabled ? <QualityAssuranceV2 /> : <QualityAssurance />;
};

export default QAFeatureToggle;
