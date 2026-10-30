// Copyright 2023 Specter Ops, Inc.
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

import { LoadingSpinner } from 'doodle-ui';
import type { FC } from 'react';

/** @deprecated Use `LoadingSpinner` from `doodle-ui` with `variant='overlay'` instead. */
const LoadingOverlay: FC<{
    loading: boolean;
}> = ({ loading }) => {
    if (!loading) return null;

    return <LoadingSpinner variant='overlay' />;
};

export default LoadingOverlay;
