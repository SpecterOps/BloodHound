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
import { Skeleton } from '@mui/material';
import { FC } from 'react';
import { AppNavigate } from '../../components';
import { useHighestPrivilegeTagId } from '../../hooks';
import { DEFAULT_ZONE_MANAGEMENT_ROUTE } from '../../routes';
const DetailsRoot: FC = () => {
    const { tagId } = useHighestPrivilegeTagId();
    if (tagId) {
        return <AppNavigate to={DEFAULT_ZONE_MANAGEMENT_ROUTE + tagId} replace />;
    } else {
        return <Skeleton className='h-24' />;
    }
};

export default DetailsRoot;
