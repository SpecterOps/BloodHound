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
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
    Dialog,
    DialogActions,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
    Input,
    Label,
    VisuallyHidden,
} from '@bloodhoundenterprise/doodleui';
import { CypherEditor } from '@neo4j-cypher/react-codemirror';
import { UpdateUserQueryRequest } from 'js-client-library';
import { useQuery } from 'react-query';
import { graphSchema } from '../../../../constants';
import { apiClient, cn } from '../../../../utils';
import { useSavedQueriesContext } from '../../providers';
import { CypherSearchState } from '../types';
import ConfirmUpdateQueryDialog from './ConfirmUpdateQueryDialog';
import SavedQueryPermissions from './SavedQueryPermissions';

const SaveQueryDialog: React.FC<{
    open: boolean;
    cypherSearchState: CypherSearchState;
    sharedIds: string[];
    isPublic: boolean;
    saveAction: string | undefined;
    saveUpdatePending: boolean;
    onClose: () => void;
    onSave: (data: { name: string; description: string; localCypherQuery: string }) => Promise<void>;
    onUpdate: (data: UpdateUserQueryRequest) => Promise<void>;
    setSharedIds: (ids: string[]) => void;
    setIsPublic: (isPublic: boolean) => void;
    error?: any;
}> = ({
    open,
    onClose,
    cypherSearchState,
    sharedIds,
    isPublic,
    saveAction,
    saveUpdatePending,
    onSave,
    onUpdate,
    setSharedIds,
    setIsPublic,
    error = undefined,
}) => {
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [id, setId] = useState<number | undefined>(undefined);
    const [isNew, setIsNew] = useState(true);
    const [localCypherQuery, setLocalCypherQuery] = useState('');
    const [isConfirmOpen, setIsConfirmOpen] = useState(false);

    const { selectedQuery } = useSavedQueriesContext();

    const { cypherQuery } = cypherSearchState;

    useEffect(() => {
        if (selectedQuery) {
            setId(selectedQuery.id);
            setIsNew(false);
        } else {
            setId(undefined);
            setIsNew(true);
        }
        setName(selectedQuery && selectedQuery.name ? selectedQuery.name : '');
        setDescription(selectedQuery && selectedQuery.description ? selectedQuery.description : '');
    }, [selectedQuery]);

    useEffect(() => {
        setLocalCypherQuery(cypherQuery);
    }, [cypherQuery]);

    const saveDisabled = name?.trim() === '' || saveUpdatePending;

    const handleSave = () => {
        if (isNew) {
            onSave({ name, description, localCypherQuery });
        } else {
            setIsConfirmOpen(true);
        }
    };
    const cypherEditorRef = useRef<CypherEditor | null>(null);
    const kindsQuery = useQuery({
        queryKey: ['graph-kinds'],
        queryFn: ({ signal }) => apiClient.getKinds({ signal }).then((res) => res.data.data.kinds),
    });

    const handleConfirmUpdate = () => {
        if (id) {
            onUpdate({ name, description, id, query: localCypherQuery });
            setIsConfirmOpen(false);
        }
    };

    const handleCancelConfirm = () => {
        setIsConfirmOpen(false);
    };

    const cardTitle =
        saveAction === 'edit' ? 'Edit Saved Query' : saveAction === 'save-as' ? 'Save As New Query' : 'Save Query';

    const dialogDescription =
        'To save your query to the Pre-built Query, add a name, optional description, and set sharing permissions.';

    return (
        <>
            <Dialog open={open} onOpenChange={(open) => !open && onClose()}>
                <DialogPortal>
                    <DialogContent className='p-0 shadow-none bg-transparent' maxWidth='lg'>
                        <VisuallyHidden>
                            <DialogTitle>{cardTitle}</DialogTitle>
                            <DialogDescription>{dialogDescription}</DialogDescription>
                        </VisuallyHidden>
                        <div className='grid grid-cols-12 gap-4' data-testId='save-query-dialog'>
                            <Card className='w-full col-span-8 p-2 rounded-lg'>
                                <CardHeader>
                                    <CardTitle>{cardTitle}</CardTitle>
                                    <CardDescription>{dialogDescription}</CardDescription>
                                </CardHeader>
                                <CardContent>
                                    <div className='mb-2'>
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

                                    <div className='mb-2'>
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
                                    <div className='mb-2 '>
                                        <Label>Cypher Query</Label>
                                        <CypherEditor
                                            ref={cypherEditorRef}
                                            className={cn(
                                                '[&_.cm-content]:saturate-200 flex grow flex-col border border-black/[.23] rounded bg-white dark:bg-[#002b36] min-h-40 max-h-40 overflow-auto [@media(min-height:720px)]:max-h-72 [&_.cm-tooltip]:max-w-lg'
                                            )}
                                            value={localCypherQuery}
                                            onValueChanged={(val: string) => {
                                                setLocalCypherQuery(val);
                                            }}
                                            theme={
                                                document.documentElement.classList.contains('dark') ? 'dark' : 'light'
                                            }
                                            schema={graphSchema(kindsQuery.data)}
                                            lineWrapping
                                            lint
                                            placeholder='Cypher Query'
                                        />
                                    </div>
                                </CardContent>
                                <CardFooter className='flex justify-end gap-4'>
                                    {error ? (
                                        <div>
                                            An error occurred while attempting to save this query. Please try again.
                                        </div>
                                    ) : null}

                                    <DialogActions className='flex justify-end gap-4'>
                                        <Button variant='text' onClick={onClose}>
                                            Cancel
                                        </Button>
                                        <Button variant='text' disabled={saveDisabled} onClick={handleSave}>
                                            Save
                                        </Button>
                                    </DialogActions>
                                </CardFooter>
                            </Card>
                            <Card className='w-full col-span-4 p-2 rounded-lg'>
                                <CardHeader>
                                    <CardTitle>Manage Shared Queries</CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <SavedQueryPermissions
                                        sharedIds={sharedIds}
                                        isPublic={isPublic}
                                        setSharedIds={setSharedIds}
                                        setIsPublic={setIsPublic}
                                    />
                                </CardContent>
                            </Card>
                        </div>
                    </DialogContent>
                </DialogPortal>
            </Dialog>
            <ConfirmUpdateQueryDialog
                handleCancel={handleCancelConfirm}
                handleApply={handleConfirmUpdate}
                open={isConfirmOpen}
                dialogContent={'Are you sure you want to update this query?'}
            />
        </>
    );
};

export default SaveQueryDialog;
