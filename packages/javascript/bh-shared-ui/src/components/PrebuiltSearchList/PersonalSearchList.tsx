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

import { Box, Skeleton, Typography } from '@mui/material';
import { FC } from 'react';
import { useNotifications } from '../../providers';
import PrebuiltSearchList, { LineItem } from './PrebuiltSearchList';
import { useDeleteSavedQuery, useSavedQueries } from '../../hooks/useSavedQueries';

// `PersonalSearchList` is a more specific implementation of `PrebuiltSearchList`.  It includes
// additional fetching logic to fetch and delete queries saved by the user
export const PersonalSearchList: FC<{ clickHandler: (query: string) => void }> = ({ clickHandler }) => {
    const userQueries = useSavedQueries();
    const deleteQueryMutation = useDeleteSavedQuery();
    const { addNotification } = useNotifications();

    const handleDeleteQuery = (id: number) =>
        deleteQueryMutation.mutate(id, {
            onSuccess: () => {
                addNotification(`Query deleted.`, 'userDeleteQuery');
            },
        });

    if (userQueries.isLoading) {
        return (
            <Box mt={2}>
                <Skeleton />
            </Box>
        );
    }

    if (userQueries.isError) {
        return (
            <Box my={2} ml={2}>
                <Typography>Unable to list saved queries.</Typography>
            </Box>
        );
    }

    const lineItems: LineItem[] =
        userQueries.data?.map((query) => ({
            description: query.name,
            cypher: query.query,
            canEdit: true,
            id: query.id,
        })) || [];

    return lineItems.length > 0 ? (
        <PrebuiltSearchList
            listSections={[
                {
                    subheader: 'User Saved Searches: ',
                    lineItems,
                },
            ]}
            clickHandler={clickHandler}
            deleteHandler={handleDeleteQuery}
        />
    ) : (
        <Box my={2} ml={2}>
            <Typography variant='body2'>No queries have been saved yet.</Typography>
        </Box>
    );
};
