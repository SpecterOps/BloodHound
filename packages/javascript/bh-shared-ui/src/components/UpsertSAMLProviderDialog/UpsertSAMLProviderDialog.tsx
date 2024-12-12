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

import { Dialog, DialogTitle } from '@mui/material';
import { SSOProvider, UpsertSAMLProviderFormInputs } from 'js-client-library';
import UpsertSAMLProviderForm from '../UpsertSAMLProviderForm';

const UpsertSAMLProviderDialog: React.FC<{
    open: boolean;
    error: any;
    oldSSOProvider?: SSOProvider;
    onClose: () => void;
    onSubmit: (data: UpsertSAMLProviderFormInputs) => void;
}> = ({ open, error, oldSSOProvider, onClose, onSubmit }) => {
    return (
        <Dialog
            open={open}
            onClose={onClose}
            fullWidth
            maxWidth='sm'
            PaperProps={{
                // @ts-ignore
                'data-testid': 'create-saml-provider-dialog',
            }}>
            <DialogTitle>{oldSSOProvider ? 'Edit' : 'Create'} SAML Provider</DialogTitle>
            <UpsertSAMLProviderForm
                error={error}
                onClose={onClose}
                oldSSOProvider={oldSSOProvider}
                onSubmit={onSubmit}
            />
        </Dialog>
    );
};

export default UpsertSAMLProviderDialog;
