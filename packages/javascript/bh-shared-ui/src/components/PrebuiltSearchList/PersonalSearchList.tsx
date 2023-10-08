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
import { FC, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { apiClient } from '../../utils';
import { useNotifications } from '../../providers';
import PrebuiltSearchList, { LineItem } from './PrebuiltSearchList';

// `PersonalSearchList` is a more specific implementation of `PrebuiltSearchList`.  It includes
// additional fetching logic to fetch and delete queries saved by the user
export const PersonalSearchList: FC<{ clickHandler: (query: string) => void }> = ({ clickHandler }) => {
    const queryClient = useQueryClient();
    const { addNotification } = useNotifications();

    const [queries, setQueries] = useState<LineItem[]>([]);

    useQuery({
        queryKey: 'userSavedQueries',
        queryFn: () => {
            return apiClient
                .getUserSavedQueries()
                .then((response) => {
                    const queries = response.data.data;

                    const queriesToDisplay = queries.map((query) => ({
                        description: query.name,
                        cypher: query.query,
                        canEdit: true,
                        id: query.id,
                    }));

                    setQueries(queriesToDisplay);
                })
                .catch(() => {
                    setQueries([]);
                });
        },
    });

    const mutation = useMutation({
        mutationFn: (queryId: number) => {
            return apiClient.deleteUserQuery(queryId);
        },
        onSettled: () => {
            queryClient.invalidateQueries({ queryKey: 'userSavedQueries' });
        },
        onSuccess: () => {
            addNotification(`Query deleted.`, 'userDeleteQuery');
        },
    });

    return queries?.length > 0 ? (
        <PrebuiltSearchList
            listSections={[
                {
                    subheader: 'User Saved Searches: ',
                    lineItems: queries,
                },
            ]}
            clickHandler={clickHandler}
            deleteHandler={mutation.mutate}
        />
    ) : (
        <Typography variant='overline' pl={'5px'} pt={'7px'} display={'block'}>
            No queries have been saved yet.
        </Typography>
    );
};
