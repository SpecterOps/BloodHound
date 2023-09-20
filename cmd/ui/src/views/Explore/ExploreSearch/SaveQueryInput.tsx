import { faSave } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button, TextField } from '@mui/material';
import { useEffect, useState } from 'react';
import { useDispatch } from 'react-redux';
import { addSnackbar } from 'src/ducks/global/actions';
import makeStyles from '@mui/styles/makeStyles';

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
    const classes = useStyles();

    const [showCustomQueryInput, setShowCustomQueryInput] = useState(false);

    const [customQueryName, setCustomQueryName] = useState('');
    const [saveDisabled, setSaveDisabled] = useState(false);

    const handleCustomQueryOnSave = () => {
        setShowCustomQueryInput((c) => !c);

        if (showCustomQueryInput) {
            setCustomQueryName(''); // reset input

            // dispatch a save to the server
            dispatch(addSnackbar(`${customQueryName} saved!`, 'customQuerySaved'));

            // dispatch a local mutation to the custom searches tab?
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
