import { Typography } from '@mui/material';
import { LineItem, PrebuiltSearchList, apiClient, useNotifications } from 'bh-shared-ui';
import { FC, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from 'react-query';

// `PersonalSearchList` is a more specific implementation of `PrebuiltSearchList`.  It includes
// additional fetching logic to fetch and delete queries saved by the user
const PersonalSearchList: FC<{ clickHandler: (query: string) => void }> = ({ clickHandler }) => {
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

export default PersonalSearchList;
