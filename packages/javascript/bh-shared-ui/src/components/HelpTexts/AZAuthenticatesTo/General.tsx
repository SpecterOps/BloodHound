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

import { FC } from 'react';

const General: FC = () => {
    return (
        <p className='edge-accordion-body2'>
            A Federated Identity Credential (FIC) is a trust configuration on an Azure App Registration that allows an
            external identity provider to authenticate as the application without a password or certificate. This edge
            indicates that the source FIC is configured on the target App Registration.
            <br />
            Any principal that can obtain a token from the FIC's trusted issuer matching its subject claim can
            authenticate as the App Registration, which in turn runs as its associated Service Principal via the
            AZRunsAs relationship.
        </p>
    );
};

export default General;
