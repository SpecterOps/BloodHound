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

import { Button, ButtonVariants } from 'doodle-ui';
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
    if (!tagId) {
        return (
            <Button variant='secondary' disabled>
                Edit {tagTypeDisplay}
            </Button>
        );
    }

    return (
        <AppLink
            data-testid='privilege-zones_edit-tag-link'
            to={tagEditLink(tagId, tagType)}
            className={ButtonVariants({
                variant: 'secondary',
            })}>
            Edit {tagTypeDisplay}
        </AppLink>
    );
};

export const CreateRuleButtonLink: FC = () => {
    const { tagId, ruleCreateLink } = usePZPathParams();
    if (!tagId) {
        return (
            <Button variant='secondary' disabled>
                Create Rule
            </Button>
        );
    }

    return (
        <AppLink
            data-testid='privilege-zones_create-rule-link'
            to={ruleCreateLink(tagId)}
            className={ButtonVariants({
                variant: 'secondary',
            })}>
            Create Rule
        </AppLink>
    );
};

export const EditRuleButtonLink: FC = () => {
    const { tagId, ruleEditLink, ruleId } = usePZPathParams();
    if (!tagId || !ruleId) {
        return (
            <Button variant='secondary' disabled>
                Edit Rule
            </Button>
        );
    }

    return (
        <AppLink
            data-testid='privilege-zones_edit-rule-link'
            to={ruleEditLink(tagId, ruleId)}
            className={ButtonVariants({
                variant: 'secondary',
            })}>
            Edit Rule
        </AppLink>
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
