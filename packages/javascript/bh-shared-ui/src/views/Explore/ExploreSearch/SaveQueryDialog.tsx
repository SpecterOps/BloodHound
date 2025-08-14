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
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
    Input,
    Label,
    VisuallyHidden,
} from '@bloodhoundenterprise/doodleui';
import { useMediaQuery, useTheme } from '@mui/material';
import { CypherEditor } from '@neo4j-cypher/react-codemirror';
import { UpdateUserQueryRequest } from 'js-client-library';
import { useQuery } from 'react-query';
import { graphSchema } from '../../../constants';
import { QueryLineItem } from '../../../types';
import { apiClient, cn } from '../../../utils';
import ConfirmUpdateQueryDialog from './ConfirmUpdateQueryDialog';
import SavedQueryPermissions from './SavedQueryPermissions';
type CypherSearchState = {
    cypherQuery: string;
    setCypherQuery: (query: string) => void;
    performSearch: (query?: string) => void;
};

const SaveQueryDialog: React.FC<{
    open: boolean;
    error?: any;
    cypherSearchState: CypherSearchState;
    selectedQuery: QueryLineItem | undefined;
    sharedIds: string[];
    isPublic: boolean;
    saveAction: string | undefined;
    onClose: () => void;
    onSave: (data: { name: string; description: string; localCypherQuery: string }) => Promise<void>;
    onUpdate: (data: UpdateUserQueryRequest) => Promise<void>;
    setSharedIds: (ids: string[]) => void;
    setIsPublic: (isPublic: boolean) => void;
}> = ({
    open,
    error = undefined,
    cypherSearchState,
    selectedQuery,
    sharedIds,
    isPublic,
    saveAction,
    onClose,
    onSave,
    onUpdate,
    setSharedIds,
    setIsPublic,
}) => {
    const theme = useTheme();

    const lgDown = useMediaQuery(theme.breakpoints.down('lg'));

    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [id, setId] = useState<number | undefined>(undefined);
    const [isNew, setIsNew] = useState(true);
    const [localCypherQuery, setLocalCypherQuery] = useState('');
    const [isConfirmOpen, setIsConfirmOpen] = useState(false);

    const { cypherQuery } = cypherSearchState;

    useEffect(() => {
        if (selectedQuery) {
            setId(selectedQuery.id);
            setIsNew(false);
        } else {
            setId(undefined);
            setIsNew(true);
        }
    }, [selectedQuery]);

    useEffect(() => {
        setName(selectedQuery && selectedQuery.name ? selectedQuery.name : '');
    }, [selectedQuery?.name]);

    useEffect(() => {
        setDescription(selectedQuery && selectedQuery.description ? selectedQuery.description : '');
    }, [selectedQuery?.description]);

    useEffect(() => {
        setLocalCypherQuery(cypherQuery);
    }, [cypherQuery]);

    const saveDisabled = name?.trim() === '';

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
        saveAction === 'edit' ? 'Edit Saved Query' : saveAction === 'saveas' ? 'Save As New Query' : 'Save Query';

    return (
        <>
            <Dialog open={open} onOpenChange={onClose}>
                <DialogPortal>
                    <DialogContent
                        DialogOverlayProps={{
                            blurBackground: false,
                        }}
                        maxWidth={lgDown ? 'md' : 'lg'}
                        className='p-0 shadow-none bg-transparent'>
                        <div className='grid grid-cols-12 gap-4'>
                            <Card className='w-full col-span-8 p-2 rounded-lg'>
                                <CardHeader>
                                    <CardTitle>{cardTitle}</CardTitle>
                                    <CardDescription>
                                        To save your query to the Pre-built Query, add a name, optional description, and
                                        set sharing permissions.
                                    </CardDescription>
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

                                    <div className='mb-2'>
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
                                            schema={graphSchema(kindsQuery.data)}
                                            lineWrapping
                                            lint
                                            placeholder='Cypher Query'
                                            tooltipAbsolute={false}
                                        />
                                    </div>
                                </CardContent>
                                <CardFooter className='flex justify-end gap-4'>
                                    {error ? (
                                        <div>
                                            An error ocurred while attempting to save this query. Please try again.
                                        </div>
                                    ) : null}

                                    <DialogActions className='flex justify-end gap-4'>
                                        <DialogClose asChild>
                                            <Button variant='text'>Cancel</Button>
                                        </DialogClose>
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
                                        queryId={selectedQuery?.id}
                                        sharedIds={sharedIds}
                                        isPublic={isPublic}
                                        setSharedIds={setSharedIds}
                                        setIsPublic={setIsPublic}
                                    />
                                </CardContent>
                            </Card>
                        </div>
                        <VisuallyHidden>
                            <DialogTitle>Save Custom Query</DialogTitle>
                            <DialogDescription>Save Custom Query</DialogDescription>
                        </VisuallyHidden>
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
