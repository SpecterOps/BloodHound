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

import { FC } from 'react';
import { useLocation } from 'react-router';

export const Create: FC = () => {
    const { state } = useLocation();

    return (
        <div>
            <h1>Create</h1>
            <h2>Type: {state.type}</h2>
            {state.within && (
                <>
                    <br />
                    Within Tier ID: {state.within}
                </>
            )}
        </div>
    );
};
