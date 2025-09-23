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

import { useQuery } from 'react-query';
import { apiClient } from '../utils';

// Named using the Minimal keyword as it uses a specific endpoint /bloodhound-users-minimal that gets active users
export const useGetUsersMinimal = () => {
    return useQuery({
        queryKey: ['users-minimal'],
        queryFn: ({ signal }) => apiClient.listUsersMinimal({ signal }).then((res) => res.data),
        select: (data) => data.data.users,
    });
};
