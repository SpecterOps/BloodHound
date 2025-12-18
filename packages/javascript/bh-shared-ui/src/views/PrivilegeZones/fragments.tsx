// Copyright 2025 Specter Ops, Inc.
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
import { FC, useContext } from 'react';
import { PrivilegeZonesContext } from './PrivilegeZonesContext';

import { Button } from '@bloodhoundenterprise/doodleui';
import { AppLink } from '../../components';
import { usePZPathParams } from '../../hooks';

const docsBasePath = 'bloodhound.specterops.io/analyze-data';
const pzPath = 'privilege-zones';

export const RulesLink: FC = () => {
    return (
        <a
            href={`https://${docsBasePath}/${pzPath}/selectors`}
            target='_blank'
            rel='noopener noreferrer'
            className='text-link underline'>
            Learn more about Rules
        </a>
    );
};

export const ZonesLink: FC = () => {
    return (
        <a
            href={`https://${docsBasePath}/${pzPath}/zones`}
            target='_blank'
            rel='noopener noreferrer'
            className='text-link underline'>
            Learn more about Zones
        </a>
    );
};

export const LabelsLink: FC = () => {
    return (
        <a
            href={`https://${docsBasePath}/${pzPath}/labels`}
            target='_blank'
            rel='noopener noreferrer'
            className='text-link underline'>
            Learn more about Labels
        </a>
    );
};

export const EditTagButtonLink: FC = () => {
    const { tagId, tagType, tagTypeDisplay, tagEditLink } = usePZPathParams();
    return (
        <Button variant='secondary' disabled={!tagId} asChild={!!tagId}>
            {!tagId ? (
                <span>Edit {tagTypeDisplay}</span>
            ) : (
                <AppLink data-testid='privilege-zones_edit-tag-link' to={tagEditLink(tagId, tagType)}>
                    Edit {tagTypeDisplay}
                </AppLink>
            )}
        </Button>
    );
};

export const CreateRuleButtonLink: FC = () => {
    const { tagId, ruleCreateLink } = usePZPathParams();
    return (
        <Button variant='secondary' disabled={!tagId} asChild={!!tagId}>
            {!tagId ? (
                <span>Create Rule</span>
            ) : (
                <AppLink data-testid='privilege-zones_create-rule-link' to={ruleCreateLink(tagId)}>
                    Create Rule
                </AppLink>
            )}
        </Button>
    );
};

export const EditRuleButtonLink: FC = () => {
    const { tagId, ruleEditLink, ruleId } = usePZPathParams();
    return (
        <Button variant='secondary' disabled={!ruleId || !tagId} asChild={!!ruleId && !!tagId}>
            {!ruleId || !tagId ? (
                <span>Edit Rule</span>
            ) : (
                <AppLink data-testid='privilege-zones_edit-rule-link' to={ruleEditLink(tagId, ruleId)}>
                    Edit Rule
                </AppLink>
            )}
        </Button>
    );
};

export const PageDescription: FC = () => {
    const { SupportLink } = useContext(PrivilegeZonesContext);

    return (
        <p className='mt-6'>
            Use Privilege Zones to segment and organize Objects based on sensitivity and access level.
            <br />
            Learn about{' '}
            <a
                href={`https://${docsBasePath}/${pzPath}/overview`}
                target='_blank'
                rel='noopener noreferrer'
                className='text-link underline'>
                setup and best practices
            </a>
            . <span>{SupportLink && <SupportLink />}</span>
        </p>
    );
};
