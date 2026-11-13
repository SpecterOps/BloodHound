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

const General: FC = () => {
    return (
        <>
            <Typography variant='body2'>The source and target nodes represent the same underlying machine.</Typography>
            <Typography variant='body2'>
                The SameHostAs edge connects different identities or representations of a machine across systems and
                platforms. For example, an Active Directory computer account, an Azure virtual machine resource, and an
                Entra device registration can all represent the same machine. These connections allow attack paths to
                cross between its identities.
            </Typography>
            <Typography variant='body2'>
                This edge describes an identity relationship. It does not grant permissions; the ability to move between
                identities depends on the access held and the machine's configuration.
            </Typography>
        </>
    );
};

export default General;
