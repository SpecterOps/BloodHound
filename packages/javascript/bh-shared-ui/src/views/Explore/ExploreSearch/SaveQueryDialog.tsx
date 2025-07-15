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

// import { Dialog, DialogActions, DialogContent, DialogTitle, FormHelperText, TextField } from '@mui/material';
import { useEffect, useRef, useState } from 'react';

import {
    Button,
    Dialog,
    DialogActions,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
    Input,
    Label,
} from '@bloodhoundenterprise/doodleui';
import { useMediaQuery, useTheme } from '@mui/material';
import { CypherEditor } from '@neo4j-cypher/react-codemirror';
import { useQuery } from 'react-query';
import { graphSchema } from '../../../constants';
import { apiClient, cn } from '../../../utils';

type CypherSearchState = {
    cypherQuery: string;
    setCypherQuery: (query: string) => void;
    performSearch: (query?: string) => void;
};

const SaveQueryDialog: React.FC<{
    open: boolean;
    onClose: () => void;
    onSave: (data: { name: string; description: string; localCypherQuery: string }) => Promise<any>;
    onUpdate: (data: {
        name: string;
        description: string;
        id: number | undefined;
        localCypherQuery: string;
    }) => Promise<any>;
    isLoading?: boolean;
    error?: any;
    cypherSearchState: CypherSearchState;
    selectedQuery: any;
}> = ({ open, onClose, onSave, onUpdate, isLoading = false, error = undefined, cypherSearchState, selectedQuery }) => {
    const theme = useTheme();

    const lgDown = useMediaQuery(theme.breakpoints.down('lg'));

    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [id, setId] = useState(undefined);
    const [isNew, setIsNew] = useState(true);
    const [localCypherQuery, setLocalCypherQuery] = useState('');
    const { cypherQuery } = cypherSearchState;

    useEffect(() => {
        if (selectedQuery) {
            //The prebuilt queries do not have a name property.  Returns undefined and throws an error surrounding controlled/uncontrolled components.  Need unified data shape for saved queries.
            setName(selectedQuery.name ? selectedQuery.name : '');
            setDescription(selectedQuery.description);
            setId(selectedQuery.id);
            setIsNew(false);
            // setLocalCypherQuery(cypherQuery);
        } else {
            setName('');
            setDescription('');
            setId(undefined);
            setIsNew(true);
            // setLocalCypherQuery(cypherQuery);
        }
    }, [selectedQuery]);

    useEffect(() => {
        setLocalCypherQuery(cypherQuery);
    }, [cypherQuery]);

    const saveDisabled = name?.trim() === '';

    const handleSave = () => {
        if (isNew) {
            onSave({ name, description, localCypherQuery });
        } else {
            onUpdate({ name, description, id, localCypherQuery });
        }
    };
    const cypherEditorRef = useRef<CypherEditor | null>(null);
    const kindsQuery = useQuery({
        queryKey: ['graph-kinds'],
        queryFn: ({ signal }) => apiClient.getKinds({ signal }).then((res) => res.data.data.kinds),
    });

    return (
        <>
            <Dialog open={open} onOpenChange={onClose}>
                <DialogPortal>
                    <DialogContent
                        DialogOverlayProps={{
                            blurBackground: false,
                        }}
                        maxWidth={lgDown ? 'sm' : 'md'}>
                        <DialogTitle>Save Query</DialogTitle>

                        <DialogDescription>
                            To save your query to the Pre-built Query, add a name, optional description, and set sharing
                            permissions.
                        </DialogDescription>

                        <div>
                            <Label htmlFor='queryName'>Query Name</Label>
                            <Input
                                type='text'
                                id='queryName'
                                value={name}
                                onChange={(e) => {
                                    setName(e.target.value);
                                }}
                            />
                        </div>

                        <div>
                            <Label htmlFor='queryDescription'>Query Description</Label>
                            <Input
                                type='text'
                                id='queryDescription'
                                value={description}
                                onChange={(e) => {
                                    setDescription(e.target.value);
                                }}
                            />
                        </div>

                        <div>
                            <Label>Cypher Query</Label>
                            <CypherEditor
                                ref={cypherEditorRef}
                                className={cn(
                                    'flex grow flex-col border border-black/[.23] rounded bg-white dark:bg-[#002b36] min-h-24 max-h-24 overflow-auto [@media(min-height:720px)]:max-h-72 [&_.cm-tooltip]:max-w-lg'
                                )}
                                value={localCypherQuery}
                                onValueChanged={(val: string) => {
                                    setLocalCypherQuery(val);
                                }}
                                theme={theme.palette.mode}
                                // onKeyDown={(e: any) => {
                                //     // if enter and shift key is pressed, execute cypher search
                                //     if (e.key === 'Enter' && e.shiftKey) {
                                //         e.preventDefault();
                                //         handleCypherSearch();
                                //     }
                                // }}
                                schema={graphSchema(kindsQuery.data)}
                                lineWrapping
                                lint
                                placeholder='Cypher Query'
                                tooltipAbsolute={false}
                            />
                        </div>

                        {error ? (
                            <div>An error ocurred while attempting to save this query. Please try again.</div>
                        ) : null}

                        <DialogActions className='flex justify-end gap-4'>
                            <DialogClose asChild>
                                <Button variant='secondary'>Cancel</Button>
                            </DialogClose>
                            <Button disabled={saveDisabled} onClick={handleSave}>
                                Save
                            </Button>
                        </DialogActions>
                    </DialogContent>
                </DialogPortal>
            </Dialog>

            {/* OLD */}

            {/* <Dialog
                open={open}
                onClose={onClose}
                disableEscapeKeyDown
                TransitionProps={{
                    onExited: () => {
                        setName('');
                    },
                }}>
                <DialogTitle>Save Query</DialogTitle>
                <DialogContent maxWidth='sm'>
                    <Input
                        type='text'
                        id='queryName'
                        value={name}
                        onChange={(e) => {
                            setName(e.target.value);
                        }}
                    />
                    {error ? (
                        <FormHelperText error>
                            An error ocurred while attempting to save this query. Please try again.
                        </FormHelperText>
                    ) : null}
                </DialogContent>
                <DialogActions>
                    <Button type='button' variant={'tertiary'} onClick={onClose} disabled={isLoading}>
                        Cancel
                    </Button>
                    <Button type='button' onClick={handleSave} disabled={saveDisabled || isLoading}>
                        Save
                    </Button>
                </DialogActions>
            </Dialog> */}
        </>
    );
};

export default SaveQueryDialog;
