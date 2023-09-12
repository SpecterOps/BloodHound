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

import { Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material';
import { Skeleton } from '@mui/material';
import SAMLProviderTableActionsMenu from '../SAMLProviderTableActionMenu';

const SAMLProviderTable: React.FC<{
    SAMLProviders: any[];
    loading: boolean;
    onDeleteSAMLProvider: (SAMLProviderId: string) => void;
}> = ({ SAMLProviders, loading, onDeleteSAMLProvider }) => {
    return (
        <Paper>
            <TableContainer>
                <Table>
                    <TableHead>
                        <TableRow>
                            <TableCell>Provider Name</TableCell>
                            <TableCell>IdP SSO URL</TableCell>
                            <TableCell>BHE SSO URL</TableCell>
                            <TableCell>BHE ACS URL</TableCell>
                            <TableCell>BHE Metadata URL</TableCell>
                            <TableCell />
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {loading ? (
                            <TableRow>
                                <TableCell>
                                    <Skeleton />
                                </TableCell>
                                <TableCell>
                                    <Skeleton />
                                </TableCell>
                                <TableCell>
                                    <Skeleton />
                                </TableCell>
                                <TableCell>
                                    <Skeleton />
                                </TableCell>
                                <TableCell>
                                    <Skeleton />
                                </TableCell>
                                <TableCell>
                                    <Skeleton />
                                </TableCell>
                            </TableRow>
                        ) : SAMLProviders.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={6} align='center'>
                                    No SAML Providers found
                                </TableCell>
                            </TableRow>
                        ) : (
                            SAMLProviders.map((SAMLProvider, i) => (
                                <TableRow key={i}>
                                    <TableCell>{SAMLProvider?.name}</TableCell>
                                    <TableCell>{SAMLProvider?.idp_sso_uri}</TableCell>
                                    <TableCell>{SAMLProvider?.sp_sso_uri}</TableCell>
                                    <TableCell>{SAMLProvider?.sp_acs_uri}</TableCell>
                                    <TableCell>{SAMLProvider?.sp_metadata_uri}</TableCell>
                                    <TableCell align='right'>
                                        <SAMLProviderTableActionsMenu
                                            SAMLProviderId={SAMLProvider.id}
                                            onDeleteSAMLProvider={onDeleteSAMLProvider}
                                        />
                                    </TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </TableContainer>
        </Paper>
    );
};
export default SAMLProviderTable;
