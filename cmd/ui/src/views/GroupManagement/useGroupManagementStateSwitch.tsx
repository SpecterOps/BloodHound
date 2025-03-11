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

import { useEnvironment, useFeatureFlag } from 'bh-shared-ui';
import { useAppSelector } from 'src/store';

const useGroupManagementStateSwitch = () => {
    const { data: flag } = useFeatureFlag('back_button_support');
    const { data: environmentFromParams } = useEnvironment();

    const reduxEnvironment = useAppSelector((state) => state.global.options.domain);
    // Nullish coalesce to null to satisfy current types. We should be able to remove this coalesce when we remove the feature flag
    const paramEnvironment = environmentFromParams ?? null;

    return flag?.enabled ? paramEnvironment : reduxEnvironment;
};

export default useGroupManagementStateSwitch;
