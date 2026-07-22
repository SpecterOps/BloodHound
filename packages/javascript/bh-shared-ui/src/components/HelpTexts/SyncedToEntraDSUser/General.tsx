// Copyright 2024 Specter Ops, Inc.
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
            <Typography variant='body2'>
                This relationship indicates that the Entra user and the Entra Domain Services user are the same identity
                across the Entra ID and managed domain boundary.
            </Typography>
            <Typography variant='body2'>
                The Entra Domain Services user is created from the Entra user during synchronization and can be
                correlated through the BloodHound aadobjectid property, collected from the LDAP attribute
                msDS-aadObjectId. Password changes in Entra ID generate and synchronize the password material required
                for the Entra Domain Services user to authenticate.
            </Typography>
            <Typography variant='body2'>
                For cloud-only users, Entra ID does not generate the NT hash required by Entra Domain Services until a
                password change occurs while the managed domain is active. A newly synchronized cloud-only user may
                exist in Entra Domain Services but remain unusable until the password is changed in Entra ID.
            </Typography>
        </>
    );
};

export default General;
