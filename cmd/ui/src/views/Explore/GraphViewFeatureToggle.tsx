import { useFeatureFlag } from 'bh-shared-ui';
import React from 'react';
import GraphView from './GraphView';
import GraphViewV2 from './GraphViewV2';

interface GraphViewFeatureToggleProps {}

const GraphViewFeatureToggle: React.FC<GraphViewFeatureToggleProps> = () => {
    const { data: flag } = useFeatureFlag('back_button_support');

    return flag?.enabled ? <GraphViewV2 /> : <GraphView />;
};

export default GraphViewFeatureToggle;
