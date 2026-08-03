// Copyright 2026 Specter Ops, Inc.
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

import { Typography } from 'doodle-ui';
import { FC } from 'react';

const Opsec: FC = () => {
    return (
        <Typography variant='body2'>
            Azure records managed-domain updates in the Activity Log. Synchronization-scope changes can also generate
            Microsoft Entra audit activity and trigger a full managed-domain resynchronization, including deletion of
            objects that fall out of scope.
        </Typography>
    );
};

export default Opsec;
