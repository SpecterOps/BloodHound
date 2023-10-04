import { LineItem, PrebuiltSearchList, apiClient } from 'bh-shared-ui';
import { FC, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { useDispatch } from 'react-redux';
import { addSnackbar } from 'src/ducks/global/actions';

// `PersonalSearchList` is a more specific implementation of `PrebuiltSearchList`.  It includes
// additional fetching logic to fetch and delete queries saved by the user
const PersonalSearchList: FC<{ clickHandler: (query: string) => void }> = ({ clickHandler }) => {
    const dispatch = useDispatch();
    const queryClient = useQueryClient();

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
            dispatch(addSnackbar(`Query deleted.`, 'userDeleteQuery'));
        },
    });

    return (
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
    );
};

export default PersonalSearchList;
