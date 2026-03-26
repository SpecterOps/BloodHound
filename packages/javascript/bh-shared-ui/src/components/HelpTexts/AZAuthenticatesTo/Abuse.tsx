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

const Abuse: FC = () => {
    return (
        <p className='edge-accordion-body2'>
            No additional abuse is necessary to traverse this edge. The abuse primitive is captured on the edge leading
            to this FIC. Once a token has been obtained from the FIC's trusted issuer, it can be exchanged at the
            Microsoft identity platform token endpoint for an access token authenticating as the target App
            Registration.
            <br />
            From there, follow the AZRunsAs edge to understand what Service Principal context, and associated
            permissions, the attacker gains.
        </p>
    );
};

export default Abuse;
