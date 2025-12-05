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
import { Button } from '@bloodhoundenterprise/doodleui';
import { AppLink, useHighestPrivilegeTagId, usePZPathParams } from 'bh-shared-ui';
import { FC } from 'react';

const InfoHeader: FC = () => {
    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const { tagId: defaultTagId, ruleCreateLink } = usePZPathParams();
    const tagId = !defaultTagId ? topTagId : defaultTagId;

    return (
        <div className='flex justify-start gap-4 items-center'>
            <Button variant='primary' disabled={!tagId} asChild={!!tagId}>
                {!tagId ? (
                    <span>Create Rule</span>
                ) : (
                    <AppLink data-testid='privilege-zones_create-rule-link' to={ruleCreateLink(tagId)}>
                        Create Rule
                    </AppLink>
                )}
            </Button>
        </div>
    );
};

export default InfoHeader;
