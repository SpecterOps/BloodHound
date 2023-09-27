import { faSave } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button, TextField } from '@mui/material';
import { FC, useEffect, useState } from 'react';
import { useDispatch } from 'react-redux';
import { addSnackbar } from 'src/ducks/global/actions';
import makeStyles from '@mui/styles/makeStyles';
import { useMutation, useQueryClient } from 'react-query';
import { apiClient } from 'bh-shared-ui';

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
    const dispatch = useDispatch();
    const queryClient = useQueryClient();

    const classes = useStyles();

    const [showCustomQueryInput, setShowCustomQueryInput] = useState(false);

    const [queryName, setQueryName] = useState('');
    const [saveButtonDisabled, setSaveButtonDisabled] = useState(false);

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
            dispatch(addSnackbar(`${queryName} saved!`, 'userSavedQuery'));
        },
    });

    const handleCustomQueryOnSave = () => {
        setShowCustomQueryInput((c) => !c);

        if (showCustomQueryInput) {
            mutation.mutate({ name: queryName, query: cypherQuery });
        }
    };

    useEffect(() => {
        const disableSaveButtonOnEmptyInput = () => {
            if (showCustomQueryInput) {
                if (!queryName) {
                    setSaveButtonDisabled(true);
                } else {
                    setSaveButtonDisabled(false);
                }
            }
        };
        disableSaveButtonOnEmptyInput();
    }, [showCustomQueryInput, queryName, setSaveButtonDisabled]);

    return (
        <>
            {showCustomQueryInput && (
                <TextField
                    label='Search Name'
                    variant='outlined'
                    size='small'
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
                onClick={handleCustomQueryOnSave}
                disabled={saveButtonDisabled}
                variant='outlined'>
                <FontAwesomeIcon icon={faSave} />
            </Button>
        </>
    );
};

export default SaveQueryInput;
