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

import { Typography } from '@mui/material';
import { FC } from 'react';

const WindowsAbuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                To abuse ownership of an object where the OWNER RIGHTS SID is explicitly granted privileges, you can
                abuse the specific privileges granted to the OWNER RIGHTS SID.
                <br />
                <br />
                Please refer to the abuse info for the specific privileges granted to OWNER RIGHTS at
                https://bloodhound.specterops.io/home/articles/17224136169371-About-BloodHound-Edges
            </Typography>
        </>
    );
};

export default WindowsAbuse;
