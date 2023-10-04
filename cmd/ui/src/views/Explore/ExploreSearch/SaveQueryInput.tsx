import { faSave } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button, TextField } from '@mui/material';
import { FC, useState } from 'react';
import makeStyles from '@mui/styles/makeStyles';
import { useMutation, useQueryClient } from 'react-query';
import { apiClient, useNotifications } from 'bh-shared-ui';

const useStyles = makeStyles((theme) => ({
    button: {
        minWidth: '35px',
        height: '35px',
        borderRadius: theme.shape.borderRadius,
        borderColor: 'rgba(0,0,0,0.23)',
        color: 'black',
    },

    iconButton: {
        padding: 0,
    },
}));

const SaveQueryInput: FC<{ cypherQuery: string }> = ({ cypherQuery }) => {
    const queryClient = useQueryClient();
    const { addNotification } = useNotifications();

    const classes = useStyles();

    const [showQueryNameInput, setShowQueryNameInput] = useState(false);
    const [queryName, setQueryName] = useState('');

    const mutation = useMutation({
        mutationFn: (newQuery: { name: string; query: string }) => {
            return apiClient.createUserQuery(newQuery);
        },
        // Always refetch after error or success:
        onSettled: () => {
            queryClient.invalidateQueries({ queryKey: 'userSavedQueries' });
        },
        onSuccess: () => {
            setQueryName(''); // reset input
            addNotification(`${queryName} saved!`, 'userSavedQuery');
        },
    });

    const handleOnSave = () => {
        setShowQueryNameInput((c) => !c);

        if (showQueryNameInput) {
            mutation.mutate({ name: queryName, query: cypherQuery });
        }
    };

    return (
        <>
            {showQueryNameInput && (
                <TextField
                    label='Search Name'
                    variant='outlined'
                    size='small'
                    fullWidth
                    sx={{ fontSize: '.875em' }}
                    InputLabelProps={{
                        style: {
                            height: '35px',
                            top: '-3px',
                        },
                    }}
                    inputProps={{
                        style: {
                            height: '35px',
                            padding: '0px 3px',
                        },
                    }}
                    value={queryName}
                    onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                        setQueryName(event.target.value);
                    }}
                />
            )}

            <Button
                className={`${classes.button} ${classes.iconButton}`}
                onClick={handleOnSave}
                disabled={showQueryNameInput && queryName.length === 0}
                variant='outlined'>
                <FontAwesomeIcon icon={faSave} />
            </Button>
        </>
    );
};

export default SaveQueryInput;
