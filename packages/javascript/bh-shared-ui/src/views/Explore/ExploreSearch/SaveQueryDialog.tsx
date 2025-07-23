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
import { useEffect, useMemo, useRef, useState } from 'react';
import { AppIcon } from '../../../components';

import {
    Button,
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
    Checkbox,
    ColumnDef,
    DataTable,
    Dialog,
    DialogActions,
    DialogClose,
    DialogContent,
    DialogPortal,
    Input,
    Label,
} from '@bloodhoundenterprise/doodleui';
import { useMediaQuery, useTheme } from '@mui/material';
import { CypherEditor } from '@neo4j-cypher/react-codemirror';
import { CheckedState } from '@radix-ui/react-checkbox';
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
    const [sharedIds, setSharedIds] = useState<string[]>([]);
    const [shareAll, setShareAll] = useState<boolean>(false);
    const [searchTerm, setSearchTerm] = useState<string>('');
    const [filteredUsers, setFilteredUsers] = useState<any[]>([]);

    const { cypherQuery } = cypherSearchState;

    const getSelf = useQuery(['getSelf'], ({ signal }) => apiClient.getSelf({ signal }).then((res) => res.data.data));

    const listUsersQuery = useQuery(['listUsers'], ({ signal }) =>
        apiClient.listUsers({ signal }).then((res) => res.data?.data?.users)
    );

    function idMap() {
        return listUsersQuery.data
            ?.filter((user: any) => user.id !== getSelf.data.id)
            .map((user: any) => {
                return {
                    isChecked: false,
                    name: user.principal_name,
                    id: user.id,
                };
            });
    }

    const usersList = useMemo(() => idMap(), [listUsersQuery.data]);
    const allUserIds = useMemo(() => usersList?.map((x) => x.id), [listUsersQuery.data]);

    useEffect(() => {
        if (selectedQuery) {
            //The prebuilt queries do not have a name property.  Returns undefined and throws an error surrounding controlled/uncontrolled components.  Need unified data shape for saved queries.
            setName(selectedQuery.name ? selectedQuery.name : '');
            setDescription(selectedQuery.description);
            setId(selectedQuery.id);
            setIsNew(false);
        } else {
            setName('');
            setDescription('');
            setId(undefined);
            setIsNew(true);
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

    const handleCheckAllChange = (checkedState: CheckedState) => {
        if (checkedState) {
            setShareAll(true);
            setSharedIds(allUserIds as string[]);
        } else {
            setShareAll(false);
            setSharedIds([]);
        }
    };

    const handleCheckChange = (id: string) => {
        if (sharedIds.includes(id)) {
            setSharedIds((prevArray) => prevArray.filter((item) => item !== id));
        } else {
            setSharedIds((prevArray) => [...prevArray, id]);
        }
    };

    const getColumns = () => {
        const columns: ColumnDef<any>[] = [
            {
                accessorKey: 'id',
                header: () => {
                    return (
                        <div className='min-w-12 max-w-12'>
                            <Checkbox className='ml-4' checked={shareAll} onCheckedChange={handleCheckAllChange} />
                        </div>
                    );
                },
                cell: ({ row }) => (
                    <>
                        <div className='min-w-12 max-w-12'>
                            <Checkbox
                                className='ml-4'
                                checked={sharedIds.includes(row.getValue('id'))}
                                onCheckedChange={() => handleCheckChange(row.getValue('id'))}
                            />
                        </div>
                    </>
                ),
            },
            {
                accessorKey: 'name',
                header: () => {
                    return <span className='dark:text-neutral-light-1'>Name</span>;
                },
                cell: ({ row }) => {
                    return (
                        <div className='dark:text-neutral-light-1 text-nowrap text-black w-full min-w-36 max-w-36'>
                            {row.getValue('name')}
                        </div>
                    );
                },
            },
        ];
        return columns;
    };

    const handleInput = (searchTerm: string) => {
        setSearchTerm(searchTerm);
    };

    useEffect(() => {
        if (searchTerm.length) {
            const filtered = usersList?.filter((user) => user.name.toLowerCase().includes(searchTerm.toLowerCase()));
            setFilteredUsers(filtered as any[]);
        } else {
            setFilteredUsers([]);
        }
    }, [searchTerm]);

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
                                    <CardTitle>Save Query</CardTitle>
                                    <CardDescription>
                                        {' '}
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
                                    {usersList && (
                                        <>
                                            <div className='flex-grow relative mb-2'>
                                                <AppIcon.MagnifyingGlass
                                                    size={16}
                                                    className='absolute left-5 top-[50%] -mt-[8px]'
                                                />
                                                <Input
                                                    type='text'
                                                    id='query-search'
                                                    placeholder='Search'
                                                    value={searchTerm}
                                                    className='w-full bg-transparent dark:bg-transparent rounded-none border-neutral-dark-5 border-t-0 border-x-0 pl-12'
                                                    onChange={(event: React.ChangeEvent<HTMLInputElement>) =>
                                                        handleInput(event.target.value)
                                                    }
                                                />
                                            </div>{' '}
                                            <DataTable
                                                TableProps={{ className: '' }}
                                                TableHeadProps={{ className: 'text-s font-bold' }}
                                                TableBodyProps={{ className: 'text-s font-roboto underline' }}
                                                TableBodyRowProps={{ className: '' }}
                                                columns={getColumns()}
                                                data={filteredUsers.length ? filteredUsers : usersList}
                                            />
                                        </>
                                    )}
                                </CardContent>
                            </Card>
                        </div>
                    </DialogContent>
                </DialogPortal>
            </Dialog>
        </>
    );
};

export default SaveQueryDialog;
