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
                This relationship indicates that BloodHound correlated a Microsoft Entra user with a user in a Microsoft
                Entra Domain Services (Entra DS) managed domain.
            </Typography>
            <Typography variant='body2'>
                Correlation uses the BloodHound aadobjectid property, collected from the LDAP attribute
                msDS-aadObjectId. Current collection does not include the Entra user&apos;s identities, creationType, or
                externalUserState properties, so B2B external identities can be misclassified. Treat this edge as
                evidence of correlation, not proof that the source user can authenticate to the managed domain.
            </Typography>
            <Typography variant='body2'>
                For cloud-only users, Entra ID does not generate the NT hash required by Entra DS until a password
                change occurs while the managed domain is active. A newly synchronized cloud-only user may exist in
                Entra DS but remain unusable until the password is changed in Entra ID. BloodHound does not verify
                password-material availability or runtime credential usability.
            </Typography>
        </>
    );
};

export default General;
