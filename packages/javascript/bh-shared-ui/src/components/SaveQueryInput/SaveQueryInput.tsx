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

import { faSave } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Button, TextField } from '@mui/material';
import { FC, useState } from 'react';
import makeStyles from '@mui/styles/makeStyles';
import { useMutation, useQueryClient } from 'react-query';
import { apiClient } from '../../utils';
import { useNotifications } from '../../providers';

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
