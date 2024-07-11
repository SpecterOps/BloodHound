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

import { useEffect, useMemo, useState } from 'react';
import {
    Box,
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Grid,
    Switch,
    Tab,
    Tabs,
    TextField,
    Typography,
    alpha,
} from '@mui/material';
import { useMutation, useQuery } from 'react-query';
import { Alert, AlertTitle } from '@mui/material';
import { PutUserAuthSecretRequest } from 'js-client-library';

import { useNotifications } from '../../providers';
import { apiClient, getUsername } from '../../utils';
import {
    Disable2FADialog,
    Enable2FADialog,
    PageWithTitle,
    PasswordDialog,
    TextWithFallback,
    UserTokenManagementDialog,
} from '../../components';

// import Box from '@mui/material/Box';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TablePagination from '@mui/material/TablePagination';
import TableRow from '@mui/material/TableRow';
import TableSortLabel from '@mui/material/TableSortLabel';
import Toolbar from '@mui/material/Toolbar';
// import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import Checkbox from '@mui/material/Checkbox';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import FormControlLabel from '@mui/material/FormControlLabel';
import { useSavedQueries } from '../../hooks';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faGlobe, faPencil, faPerson, faSquareMinus } from '@fortawesome/free-solid-svg-icons';
// import Switch from '@mui/material/Switch';

const UserProfile = () => {
    const { addNotification } = useNotifications();
    const [changePasswordDialogOpen, setChangePasswordDialogOpen] = useState(false);
    const [userTokenManagementDialogOpen, setUserTokenManagementDialogOpen] = useState(false);
    const [enable2FADialogOpen, setEnable2FADialogOpen] = useState(false);
    const [disable2FADialogOpen, setDisable2FADialogOpen] = useState(false);
    const [TOTPSecret, setTOTPSecret] = useState('');
    const [QRCode, setQRCode] = useState('');
    const [enable2FAError, setEnable2FAError] = useState('');
    const [disable2FAError, setDisable2FAError] = useState('');
    const [disable2FASecret, setDisable2FASecret] = useState('');

    // handle table changing
    const [value, setValue] = useState(0);

    const handleChange = (event: React.SyntheticEvent, newValue: number) => {
        setValue(newValue);
    };

    const getSelfQuery = useQuery(['getSelf'], ({ signal }) =>
        apiClient.getSelf({ signal }).then((res) => res.data.data)
    );

    const updateUserPasswordMutation = useMutation(
        ({ userId, ...payload }: { userId: string } & PutUserAuthSecretRequest) =>
            apiClient.putUserAuthSecret(userId, payload),
        {
            onSuccess: () => {
                addNotification('Password updated successfully!', 'updateUserPasswordSuccess');
                setChangePasswordDialogOpen(false);
            },
            onError: (error: any) => {
                if (error.response?.status == 403) {
                    addNotification(
                        'Current password invalid. Password update failed.',
                        'UpdateUserPasswordCurrentPasswordInvalidError'
                    );
                } else {
                    addNotification('Password failed to update.', 'UpdateUserPasswordError');
                }
            },
        }
    );

    if (getSelfQuery.isLoading) {
        return (
            <PageWithTitle
                title='My Profile'
                data-testid='my-profile'
                pageDescription={
                    <Typography variant='body2' paragraph>
                        Review and manage your user account.
                    </Typography>
                }>
                <Typography variant='h2'>User Information</Typography>
                <Box p={4} textAlign='center'>
                    <CircularProgress />
                </Box>
            </PageWithTitle>
        );
    }

    if (getSelfQuery.isError) {
        return (
            <PageWithTitle
                title='My Profile'
                data-testid='my-profile'
                pageDescription={
                    <Typography variant='body2' paragraph>
                        Review and manage your user account.
                    </Typography>
                }>
                <Typography variant='h2'>User Information</Typography>

                <Alert severity='error'>
                    <AlertTitle>Error</AlertTitle>
                    Sorry, there was a problem fetching your user information.
                    <br />
                    Please try refreshing the page or logging in again.
                </Alert>
            </PageWithTitle>
        );
    }

    const user = getSelfQuery.data;

    return (
        <>
            <PageWithTitle
                title='My Profile'
                data-testid='my-profile'
                pageDescription={
                    <Typography variant='body2' paragraph>
                        Review and manage your user account.
                    </Typography>
                }>
                <Tabs value={value} onChange={handleChange}>
                    <Tab label='Profile Information' />
                    <Tab label='Custom Searches' />
                </Tabs>

                <CustomTabPanel index={0} value={value}>
                    <Typography variant='h2'>User Information</Typography>

                    <Grid container spacing={2} alignItems='center'>
                        <Grid item xs={3}>
                            <Typography variant='body1'>Email</Typography>
                        </Grid>
                        <Grid item xs={9}>
                            <Typography variant='body1'>{user?.email_address}</Typography>
                        </Grid>

                        <Grid item xs={3}>
                            <Typography variant='body1'>Name</Typography>
                        </Grid>
                        <Grid item xs={9}>
                            <Typography variant='body1'>
                                <TextWithFallback text={getUsername(user)} fallback='Unknown' />
                            </Typography>
                        </Grid>

                        <Grid item xs={3}>
                            <Typography variant='body1'>Role</Typography>
                        </Grid>
                        <Grid item xs={9}>
                            <Typography variant='body1'>
                                <TextWithFallback text={user?.roles?.[0]?.name} fallback='Unknown' />
                            </Typography>
                        </Grid>
                    </Grid>

                    <Box mt={2}>
                        <Typography variant='h2'>Authentication</Typography>
                    </Box>
                    <Grid container spacing={2} alignItems='center'>
                        <Grid container item>
                            <Grid item xs={3}>
                                <Typography variant='body1'>API Key Management</Typography>
                            </Grid>
                            <Grid item xs={2}>
                                <Button
                                    variant='contained'
                                    color='primary'
                                    size='small'
                                    disableElevation
                                    fullWidth
                                    onClick={() => setUserTokenManagementDialogOpen(true)}
                                    data-testid='my-profile_button-api-key-management'>
                                    API Key Management
                                </Button>
                            </Grid>
                        </Grid>
                        {user.saml_provider_id === null && (
                            <>
                                <Grid container item>
                                    <Grid item xs={3}>
                                        <Typography variant='body1'>Password</Typography>
                                    </Grid>
                                    <Grid item xs={2}>
                                        <Button
                                            variant='contained'
                                            color='primary'
                                            size='small'
                                            disableElevation
                                            fullWidth
                                            onClick={() => setChangePasswordDialogOpen(true)}
                                            data-testid='my-profile_button-reset-password'>
                                            Reset Password
                                        </Button>
                                    </Grid>
                                </Grid>

                                <Grid container item>
                                    <Grid item xs={3}>
                                        <Typography variant='body1'>Multi-Factor Authentication</Typography>
                                    </Grid>
                                    <Grid item xs={9}>
                                        <Box display='flex' alignItems='center'>
                                            <Switch
                                                inputProps={{
                                                    'aria-label': 'Multi-Factor Authentication Enabled',
                                                }}
                                                checked={user.AuthSecret?.totp_activated}
                                                onChange={() => {
                                                    if (!user.AuthSecret?.totp_activated) setEnable2FADialogOpen(true);
                                                    else setDisable2FADialogOpen(true);
                                                }}
                                                color='primary'
                                                data-testid='my-profile_switch-multi-factor-authentication'
                                            />
                                            {user.AuthSecret?.totp_activated && (
                                                <Typography variant='body1'>Enabled</Typography>
                                            )}
                                        </Box>
                                    </Grid>
                                </Grid>
                            </>
                        )}
                    </Grid>
                </CustomTabPanel>

                <CustomTabPanel index={1} value={value}>
                    <EnhancedTable />
                </CustomTabPanel>
            </PageWithTitle>

            <PasswordDialog
                open={changePasswordDialogOpen}
                onClose={() => setChangePasswordDialogOpen(false)}
                userId={user.id}
                requireCurrentPassword={true}
                showNeedsPasswordReset={false}
                onSave={updateUserPasswordMutation.mutate}
            />

            <UserTokenManagementDialog
                open={userTokenManagementDialogOpen}
                onClose={() => setUserTokenManagementDialogOpen(false)}
                userId={user.id}
            />

            <Enable2FADialog
                open={enable2FADialogOpen}
                onClose={() => {
                    setEnable2FADialogOpen(false);
                    setEnable2FAError('');
                    setDisable2FASecret('');
                    getSelfQuery.refetch();
                }}
                onCancel={() => {
                    setEnable2FADialogOpen(false);
                    setEnable2FAError('');
                    setDisable2FASecret('');
                    getSelfQuery.refetch();
                }}
                onSavePassword={(password) => {
                    setEnable2FAError('');
                    return apiClient
                        .enrollMFA(user.id, {
                            secret: password,
                        })
                        .then((response) => {
                            setQRCode(response.data.data.qr_code);
                            setTOTPSecret(response.data.data.totp_secret);
                            setEnable2FAError('');
                        })
                        .catch((err) => {
                            setEnable2FAError('Unable to verify password. Please try again.');
                            throw err;
                        });
                }}
                onSaveOTP={(OTP) => {
                    setEnable2FAError('');
                    return apiClient
                        .activateMFA(user.id, {
                            otp: OTP,
                        })
                        .then(() => {
                            setEnable2FAError('');
                        })
                        .catch((err) => {
                            setEnable2FAError('Unable to verify one-time password. Please try again.');
                            throw err;
                        });
                }}
                onSave={() => {
                    setEnable2FADialogOpen(false);
                    setEnable2FAError('');
                    getSelfQuery.refetch();
                }}
                TOTPSecret={TOTPSecret}
                QRCode={QRCode}
                error={enable2FAError}
            />

            <Disable2FADialog
                open={disable2FADialogOpen}
                onClose={() => {
                    setDisable2FADialogOpen(false);
                    setDisable2FAError('');
                    getSelfQuery.refetch();
                }}
                onCancel={() => {
                    setDisable2FADialogOpen(false);
                    setDisable2FAError('');
                    getSelfQuery.refetch();
                }}
                onSave={(secret: string) => {
                    setDisable2FAError('');
                    apiClient
                        .disenrollMFA(user.id, { secret })
                        .then(() => {
                            setDisable2FADialogOpen(false);
                            setDisable2FAError('');
                            setDisable2FASecret('');
                            getSelfQuery.refetch();
                        })
                        .catch(() => {
                            setDisable2FAError('Unable to verify password. Please try again.');
                        });
                }}
                error={disable2FAError}
                secret={disable2FASecret}
                onSecretChange={(e: any) => setDisable2FASecret(e.target.value)}
                contentText='To stop using multi-factor authentication, please enter your password for security purposes.'
            />
        </>
    );
};

interface TabPanelProps {
    children?: React.ReactNode;
    index: number;
    value: number;
}

function CustomTabPanel(props: TabPanelProps) {
    const { children, value, index, ...other } = props;

    return (
        <div
            role='tabpanel'
            hidden={value !== index}
            id={`simple-tabpanel-${index}`}
            aria-labelledby={`simple-tab-${index}`}
            {...other}>
            {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
        </div>
    );
}

export default UserProfile;

interface Data {
    id: number;
    scope: string;
    name: string;
    description: string;
    permissions: boolean;
    query: string;
}

function createData(
    id: number,
    scope: string,
    name: string,
    description: string,
    permissions: boolean,
    query: string
): Data {
    return {
        id,
        name,
        scope,
        description,
        permissions,
        query,
    };
}

const rows = [
    createData(1, 'global', 'custom query 1', 'description 1', false, 'match (x) return x'),
    createData(2, 'global', 'custom query 2', 'description 2', false, 'match (x) return x'),
    createData(3, 'global', 'custom query 3', 'description 3', false, 'match (x) return x'),
    createData(4, 'shared', 'custom query 4', 'description 4', true, 'match (x) return x'),
    createData(5, 'shared', 'custom query 5', 'description 5', true, 'match (x) return x'),
    createData(6, 'personal', 'custom query 6', 'description 6', true, 'match (x) return x'),
];

function descendingComparator<T>(a: T, b: T, orderBy: keyof T) {
    if (b[orderBy] < a[orderBy]) {
        return -1;
    }
    if (b[orderBy] > a[orderBy]) {
        return 1;
    }
    return 0;
}

type Order = 'asc' | 'desc';

function getComparator<Key extends keyof any>(
    order: Order,
    orderBy: Key
): (a: { [key in Key]: number | string }, b: { [key in Key]: number | string }) => number {
    return order === 'desc'
        ? (a, b) => descendingComparator(a, b, orderBy)
        : (a, b) => -descendingComparator(a, b, orderBy);
}

// Since 2020 all major browsers ensure sort stability with Array.prototype.sort().
// stableSort() brings sort stability to non-modern browsers (notably IE11). If you
// only support modern browsers you can replace stableSort(exampleArray, exampleComparator)
// with exampleArray.slice().sort(exampleComparator)
function stableSort<T>(array: readonly T[], comparator: (a: T, b: T) => number) {
    const stabilizedThis = array.map((el, index) => [el, index] as [T, number]);
    stabilizedThis.sort((a, b) => {
        const order = comparator(a[0], b[0]);
        if (order !== 0) {
            return order;
        }
        return a[1] - b[1];
    });
    return stabilizedThis.map((el) => el[0]);
}

interface HeadCell {
    disablePadding: boolean;
    id: keyof Data;
    label: string;
    numeric: boolean;
}

const headCells: readonly HeadCell[] = [
    {
        id: 'scope',
        numeric: false,
        disablePadding: true,
        label: 'Scope',
    },
    {
        id: 'name',
        numeric: false,
        disablePadding: false,
        label: 'Name',
    },
    {
        id: 'description',
        numeric: false,
        disablePadding: false,
        label: 'Description',
    },
    {
        id: 'permissions',
        numeric: false,
        disablePadding: false,
        label: 'Permissions',
    },
];

interface EnhancedTableProps {
    numSelected: number;
    onRequestSort: (event: React.MouseEvent<unknown>, property: keyof Data) => void;
    onSelectAllClick: (event: React.ChangeEvent<HTMLInputElement>) => void;
    order: Order;
    orderBy: string;
    rowCount: number;
}

function EnhancedTableHead(props: EnhancedTableProps) {
    const { onSelectAllClick, order, orderBy, numSelected, rowCount, onRequestSort } = props;
    const createSortHandler = (property: keyof Data) => (event: React.MouseEvent<unknown>) => {
        onRequestSort(event, property);
    };

    return (
        <TableHead>
            <TableRow>
                {/* <TableCell padding='checkbox'>
                    <Checkbox
                        color='primary'
                        indeterminate={numSelected > 0 && numSelected < rowCount}
                        checked={rowCount > 0 && numSelected === rowCount}
                        onChange={onSelectAllClick}
                        inputProps={{
                            'aria-label': 'select all desserts',
                        }}
                    />
                </TableCell> */}
                {headCells.map((headCell) => (
                    <TableCell
                        key={headCell.id}
                        align={headCell.numeric ? 'right' : 'left'}
                        padding={headCell.disablePadding ? 'none' : 'normal'}
                        sortDirection={orderBy === headCell.id ? order : false}>
                        <TableSortLabel
                            active={orderBy === headCell.id}
                            direction={orderBy === headCell.id ? order : 'asc'}
                            onClick={createSortHandler(headCell.id)}>
                            {headCell.label}
                            {/* {orderBy === headCell.id ? (
                                <Box component='span'>
                                    {order === 'desc' ? 'sorted descending' : 'sorted ascending'}
                                </Box>
                            ) : null} */}
                        </TableSortLabel>
                    </TableCell>
                ))}
            </TableRow>
        </TableHead>
    );
}

interface EnhancedTableToolbarProps {
    numSelected: number;
}

function EnhancedTableToolbar(props: EnhancedTableToolbarProps) {
    const { numSelected } = props;

    return (
        <Toolbar
            sx={{
                pl: { sm: 2 },
                pr: { xs: 1, sm: 1 },
                ...(numSelected > 0 && {
                    bgcolor: (theme) => alpha(theme.palette.primary.main, theme.palette.action.activatedOpacity),
                }),
            }}>
            {numSelected > 0 ? (
                <Typography sx={{ flex: '1 1 100%' }} color='inherit' variant='subtitle1' component='div'>
                    {numSelected} selected
                </Typography>
            ) : (
                <Typography sx={{ flex: '1 1 100%' }} variant='h6' id='tableTitle' component='div'>
                    Manage Custom Searches
                </Typography>
            )}
            {numSelected > 0 ? (
                <Tooltip title='Delete'>
                    <IconButton></IconButton>
                </Tooltip>
            ) : (
                <Tooltip title='Filter list'>
                    <IconButton></IconButton>
                </Tooltip>
            )}
        </Toolbar>
    );
}

export function EnhancedTable() {
    const [order, setOrder] = useState<Order>('asc');
    const [orderBy, setOrderBy] = useState<keyof Data>('scope');
    const [selected, setSelected] = useState<readonly number[]>([]);
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(15);

    const [openDialog, setOpenDialog] = useState(false);

    const userQueries = useSavedQueries();
    console.log(userQueries.data);

    useEffect(() => {
        if (selected) {
            console.log(selected);
        }
    }, [selected]);

    const handleRequestSort = (event: React.MouseEvent<unknown>, property: keyof Data) => {
        const isAsc = orderBy === property && order === 'asc';
        setOrder(isAsc ? 'desc' : 'asc');
        setOrderBy(property);
    };

    const handleSelectAllClick = (event: React.ChangeEvent<HTMLInputElement>) => {
        if (event.target.checked) {
            const newSelected = rows.map((n) => n.id);
            setSelected(newSelected);
            return;
        }
        setSelected([]);
    };

    const handleClick = (event: React.MouseEvent<unknown>, id: number) => {
        const selectedIndex = selected.indexOf(id);
        let newSelected: readonly number[] = [];

        if (selectedIndex === -1) {
            newSelected = newSelected.concat(selected, id);
        } else if (selectedIndex === 0) {
            newSelected = newSelected.concat(selected.slice(1));
        } else if (selectedIndex === selected.length - 1) {
            newSelected = newSelected.concat(selected.slice(0, -1));
        } else if (selectedIndex > 0) {
            newSelected = newSelected.concat(selected.slice(0, selectedIndex), selected.slice(selectedIndex + 1));
        }
        setSelected(newSelected);
    };

    const handleChangePage = (event: unknown, newPage: number) => {
        setPage(newPage);
    };

    const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
        setRowsPerPage(parseInt(event.target.value, 10));
        setPage(0);
    };

    const isSelected = (id: number) => selected.indexOf(id) !== -1;

    // Avoid a layout jump when reaching the last page with empty rows.
    const emptyRows = page > 0 ? Math.max(0, (1 + page) * rowsPerPage - rows.length) : 0;

    const visibleRows = useMemo(
        () =>
            stableSort(rows, getComparator(order, orderBy)).slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage),
        [order, orderBy, page, rowsPerPage]
    );

    const handleClose = () => {
        setOpenDialog(false);
    };

    const handleSaveEdits = () => {};

    const handleDelete = () => {};

    return (
        <Box sx={{ width: '100%' }}>
            <Paper sx={{ width: '100%', mb: 2 }}>
                <EnhancedTableToolbar numSelected={selected.length} />
                <TableContainer>
                    <Table sx={{ minWidth: 750 }} aria-labelledby='tableTitle'>
                        <EnhancedTableHead
                            numSelected={selected.length}
                            order={order}
                            orderBy={orderBy}
                            onSelectAllClick={handleSelectAllClick}
                            onRequestSort={handleRequestSort}
                            rowCount={rows.length}
                        />
                        <TableBody>
                            {visibleRows.map((row, index) => {
                                const isItemSelected = isSelected(row.id);
                                const labelId = `enhanced-table-checkbox-${index}`;

                                return (
                                    <TableRow
                                        hover
                                        onClick={(event) => handleClick(event, row.id)}
                                        role='checkbox'
                                        aria-checked={isItemSelected}
                                        tabIndex={-1}
                                        key={row.id}
                                        selected={isItemSelected}
                                        sx={{ cursor: 'pointer' }}>
                                        <TableCell component='th' id={labelId} scope='row'>
                                            {row.scope === 'global' ? (
                                                <FontAwesomeIcon icon={faGlobe} />
                                            ) : row.scope === 'shared' ? (
                                                <>
                                                    <FontAwesomeIcon icon={faPerson} />
                                                    <FontAwesomeIcon icon={faPerson} />
                                                </>
                                            ) : (
                                                <FontAwesomeIcon icon={faPerson} />
                                            )}
                                        </TableCell>
                                        <TableCell align='left'>{row.name}</TableCell>
                                        <TableCell align='left'>{row.description}</TableCell>
                                        <TableCell
                                            align='left'
                                            onClick={() => {
                                                if (visibleRows[index].permissions) {
                                                    setOpenDialog(true);
                                                }
                                            }}>
                                            {row.permissions ? (
                                                <FontAwesomeIcon icon={faPencil} />
                                            ) : (
                                                <FontAwesomeIcon icon={faSquareMinus} />
                                            )}
                                        </TableCell>
                                    </TableRow>
                                );
                            })}
                            {emptyRows > 0 && (
                                <TableRow
                                    style={{
                                        height: 53 * emptyRows,
                                    }}>
                                    <TableCell colSpan={6} />
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </TableContainer>
                <TablePagination
                    rowsPerPageOptions={[5, 10, 25]}
                    component='div'
                    count={rows.length}
                    rowsPerPage={rowsPerPage}
                    page={page}
                    onPageChange={handleChangePage}
                    onRowsPerPageChange={handleChangeRowsPerPage}
                />
            </Paper>

            {selected.length === 1 && (
                <Dialog open={openDialog} maxWidth={'md'} fullWidth>
                    <DialogTitle>Edit Custom Search</DialogTitle>
                    <DialogContent>
                        <ul>
                            <li>
                                <TextField
                                    id='standard-basic'
                                    label='Search Name'
                                    variant='standard'
                                    defaultValue={visibleRows.find((row) => row.id === selected[0]).name || undefined}
                                    sx={{ width: '500px' }}
                                />
                            </li>
                            <li>
                                <TextField
                                    id='standard-basic'
                                    label='Search Description'
                                    variant='standard'
                                    defaultValue={
                                        visibleRows.find((row) => row.id === selected[0])?.description || undefined
                                    }
                                    sx={{ width: '500px' }}
                                />
                            </li>
                            <li>
                                <TextField
                                    id='standard-basic'
                                    label='Share With'
                                    variant='standard'
                                    sx={{ width: '500px' }}
                                />
                            </li>
                            <li>
                                <TextField
                                    id='standard-multiline-static'
                                    label='Cypher Query'
                                    multiline
                                    rows={4}
                                    variant='standard'
                                    defaultValue={visibleRows.find((row) => row.id === selected[0]).query || undefined}
                                    sx={{ width: '500px' }}
                                />
                            </li>
                        </ul>
                    </DialogContent>
                    <DialogActions>
                        <Box display={'flex'} justifyContent={'space-between'} width={'100%'}>
                            <div>
                                <Button autoFocus onClick={handleClose}>
                                    delete
                                </Button>
                            </div>
                            <div>
                                <Button autoFocus onClick={handleClose}>
                                    close
                                </Button>
                                <Button onClick={handleSaveEdits}>Ok</Button>
                            </div>
                        </Box>
                    </DialogActions>
                </Dialog>
            )}
        </Box>
    );
}
