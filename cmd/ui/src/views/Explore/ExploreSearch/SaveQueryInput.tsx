import { faSave } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button, TextField } from '@mui/material';
import { useEffect, useState } from 'react';
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

const SaveQueryInput = () => {
    const dispatch = useDispatch();
    const queryClient = useQueryClient();

    const classes = useStyles();

    const [showCustomQueryInput, setShowCustomQueryInput] = useState(false);

    const [customQueryName, setCustomQueryName] = useState('');
    const [saveDisabled, setSaveDisabled] = useState(false);

    const mutation = useMutation({
        mutationFn: (newQuery: any) => {
            return apiClient.createUserQuery(newQuery);
        },
        // Always refetch after error or success:
        onSettled: () => {
            queryClient.invalidateQueries({ queryKey: 'userSavedQueries' });
        },
        onSuccess: () => {
            setCustomQueryName(''); // reset input
            dispatch(addSnackbar(`${customQueryName} saved!`, 'userSavedQuery'));
        },
    });

    const handleCustomQueryOnSave = () => {
        setShowCustomQueryInput((c) => !c);

        if (showCustomQueryInput) {
            mutation.mutate({ name: customQueryName, query: 'match (n) return n limit 100' });
        }
    };

    useEffect(() => {
        const disableSaveButtonOnEmptyInput = () => {
            if (showCustomQueryInput) {
                if (!customQueryName) {
                    setSaveDisabled(true);
                } else {
                    setSaveDisabled(false);
                }
            }
        };
        disableSaveButtonOnEmptyInput();
    }, [showCustomQueryInput, customQueryName, setSaveDisabled]);

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
                            // ...(!focused && { top: `${labelOffset}px` }),
                        },
                    }}
                    inputProps={{
                        style: {
                            height: '35px',
                            padding: '0px 3px',
                        },
                    }}
                    value={customQueryName}
                    onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                        setCustomQueryName(event.target.value);
                    }}
                />
            )}

            <Button
                className={`${classes.button} ${classes.iconButton}`}
                onClick={handleCustomQueryOnSave}
                disabled={saveDisabled}
                variant='outlined'>
                <FontAwesomeIcon icon={faSave} />
            </Button>
        </>
    );
};

export default SaveQueryInput;
