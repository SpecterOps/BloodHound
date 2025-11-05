// Copyright 2024 Specter Ops, Inc.
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

import { Button, Label } from '@bloodhoundenterprise/doodleui';
import { OIDCProviderInfo, Role, SAMLProviderInfo, SSOProvider } from 'js-client-library';
import fileDownload from 'js-file-download';
import { FC, useMemo } from 'react';
import { useNotifications } from '../../providers';
import { apiClient } from '../../utils';
import { Field, FieldsContainer } from '../../views/Explore/fragments';
import LabelWithCopy from '../LabelWithCopy';

const SAMLProviderInfoPanel: FC<{
    samlProviderDetails: SAMLProviderInfo;
}> = ({ samlProviderDetails }) => (
    <>
        <Field
            label={<LabelWithCopy label='IdP SSO URL' valueToCopy={samlProviderDetails.idp_sso_uri} hoverOnly />}
            value={samlProviderDetails.idp_sso_uri}
        />
        <Field
            label={<LabelWithCopy label='BHE SSO URL' valueToCopy={samlProviderDetails.sp_sso_uri} hoverOnly />}
            value={samlProviderDetails.sp_sso_uri}
        />
        <Field
            label={<LabelWithCopy label='BHE ACS URL' valueToCopy={samlProviderDetails.sp_acs_uri} hoverOnly />}
            value={samlProviderDetails.sp_acs_uri}
        />
        <Field
            label={
                <LabelWithCopy label='BHE Metadata URL' valueToCopy={samlProviderDetails.sp_metadata_uri} hoverOnly />
            }
            value={samlProviderDetails.sp_metadata_uri}
        />
    </>
);

const OIDCProviderInfoPanel: FC<{
    ssoProvider: SSOProvider;
}> = ({ ssoProvider }) => {
    const oidcProviderDetails = ssoProvider.details as OIDCProviderInfo;
    return (
        <>
            <Field
                label={<LabelWithCopy label='Client ID' valueToCopy={oidcProviderDetails.client_id} hoverOnly />}
                value={oidcProviderDetails.client_id}
            />
            <Field
                label={<LabelWithCopy label='Issuer' valueToCopy={oidcProviderDetails.issuer} hoverOnly />}
                value={oidcProviderDetails.issuer}
            />
            <Field
                label={<LabelWithCopy label='Callback URL' valueToCopy={ssoProvider.callback_uri} hoverOnly />}
                value={ssoProvider.callback_uri}
            />
        </>
    );
};

const SSOProviderInfoPanel: FC<{
    ssoProvider: SSOProvider;
    roles?: Role[];
}> = ({ ssoProvider, roles }) => {
    const { addNotification } = useNotifications();

    const defaultRoleName = useMemo(
        () => roles?.find((role) => role.id === ssoProvider.config?.auto_provision?.default_role_id)?.name,
        [roles, ssoProvider.config?.auto_provision?.default_role_id]
    );

    if (!ssoProvider.type) {
        return null;
    }

    let innerInfoPanel;
    switch (ssoProvider.type.toLowerCase()) {
        case 'saml':
            innerInfoPanel = <SAMLProviderInfoPanel samlProviderDetails={ssoProvider.details as SAMLProviderInfo} />;
            break;
        case 'oidc':
            innerInfoPanel = <OIDCProviderInfoPanel ssoProvider={ssoProvider} />;
            break;
        default:
            innerInfoPanel = null;
    }

    const downloadSAMLSigningCertificate = () => {
        if (ssoProvider.type.toLowerCase() == 'oidc') {
            addNotification('Only SAML providers support signing certificates.', 'errorDownloadSAMLSigningCertificate');
        } else {
            apiClient
                .getSAMLProviderSigningCertificate(ssoProvider.id)
                .then((res) => {
                    const filename =
                        res.headers['content-disposition']?.match(/^.*filename="(.*)"$/)?.[1] ||
                        `${ssoProvider.name}-signing-certificate`;

                    fileDownload(res.data, filename);
                })
                .catch((err) => {
                    console.error(err);
                    addNotification(
                        'This file could not be downloaded. Please try again.',
                        'downloadSAMLSigningCertificate'
                    );
                });
        }
    };

    return (
        <div className='width-[400px] max-w-[400px]'>
            <div className='flex flex-col overflow-y-hidden h-full' data-testid='sso_provider-info-panel'>
                <div>
                    <div className='flex items-center bg-neutral-5'>
                        <div className='bg-primary w-2 h-14 mr-2'></div>
                        <h5
                            data-testid='sso_provider-info-panel_header-text'
                            className='text-nowrap grow text-lg font-bold'>
                            {ssoProvider?.name}
                        </h5>
                    </div>
                    <div className='bg-neutral-2 overflow-x-hidden overflow-y-auto px-4 py-2 shadow-outer-1 rounded'>
                        <div className='shrink-0 grow font-bold ml-2 text-sm'>Provider Information:</div>
                        <FieldsContainer>
                            {innerInfoPanel}
                            <Field
                                label={<Label className='text-xs'>Automatically create new users on login</Label>}
                                value={ssoProvider.config?.auto_provision?.enabled ? 'Yes' : 'No'}
                            />
                            {ssoProvider.config?.auto_provision?.enabled && (
                                <>
                                    <Field
                                        label={
                                            <Label className='text-xs'>
                                                Allow SSO provider to manage roles for new users
                                            </Label>
                                        }
                                        value={ssoProvider.config?.auto_provision?.role_provision ? 'Yes' : 'No'}
                                    />
                                    <Field
                                        label={<Label className='text-xs'>Default role when creating new users</Label>}
                                        value={defaultRoleName ?? 'Read-Only'}
                                    />
                                </>
                            )}
                        </FieldsContainer>
                    </div>
                </div>
            </div>
            {ssoProvider.type.toLowerCase() === 'saml' && (
                <div className='flex justify-center mt-2'>
                    <Button
                        aria-label={`Download ${ssoProvider.name} SP Certificate`}
                        variant='secondary'
                        onClick={downloadSAMLSigningCertificate}>
                        Download SAML SP Certificate
                    </Button>
                </div>
            )}
        </div>
    );
};

export default SSOProviderInfoPanel;
