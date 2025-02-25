// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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
