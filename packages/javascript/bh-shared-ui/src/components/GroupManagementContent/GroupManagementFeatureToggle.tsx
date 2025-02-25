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

import React from 'react';
import { useFeatureFlag } from '../../hooks/useFeatureFlags';
import GroupManagementContent from './GroupManagementContent';
import GroupManagementContentV2, { GroupManagementContentV2Props } from './GroupManagementContentV2';

const GroupManagementFeatureToggle: React.FC<GroupManagementContentV2Props> = (props) => {
    const { data: flag } = useFeatureFlag('back_button_support');
    return flag?.enabled ? <GroupManagementContentV2 {...props} /> : <GroupManagementContent {...props} />;
};

export default GroupManagementFeatureToggle;
